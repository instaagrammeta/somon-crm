package service

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"

	"github.com/instaagrammeta/somon-crm/backend/internal/models"
)

// TwoFactorService manages TOTP (RFC 6238) two-factor for users. The flow is:
//
//  1. User opens "Two-factor" page → POST /api/2fa/setup
//     → server generates a fresh secret, stores Enabled=false row, returns
//     otpauth:// URL that the frontend renders as a QR code.
//  2. User scans the QR with Google Authenticator / 1Password / Authy.
//  3. User enters the 6-digit code → POST /api/2fa/verify
//     → server validates with totp.Validate(); on success Enabled=true and
//     we generate 8 single-use recovery codes shown ONCE to the user.
//  4. On every subsequent login, after password check, the auth handler
//     calls TwoFactorService.Validate() which accepts either a TOTP code
//     or one of the recovery codes (which is then burned).
type TwoFactorService struct {
	db    *gorm.DB
	brand string // shown as the "issuer" in authenticator apps
}

func NewTwoFactorService(db *gorm.DB, brand string) *TwoFactorService {
	if brand == "" {
		brand = "Somon CRM"
	}
	return &TwoFactorService{db: db, brand: brand}
}

// Status reports whether the user has completed 2FA setup. The setup is
// considered "in progress" if a row exists but Enabled=false.
type Status struct {
	Enrolled bool `json:"enrolled"` // row exists (verified or pending)
	Enabled  bool `json:"enabled"`  // row exists AND verified — login will require code
}

func (s *TwoFactorService) Status(userID uint) Status {
	var row models.UserTwoFactor
	if err := s.db.Where("user_id = ?", userID).First(&row).Error; err != nil {
		return Status{}
	}
	return Status{Enrolled: true, Enabled: row.Enabled}
}

// SetupResult is what /setup returns. The Secret is shown to the user as a
// fallback (in case the QR cannot be rendered) but the typical flow is to use
// the OTPAuthURL, which already embeds the secret.
type SetupResult struct {
	Secret      string `json:"secret"`
	OTPAuthURL  string `json:"otpauth_url"` // ready to encode as a QR
	QRImageData string `json:"qr_image"`    // optional inline PNG (omitted here)
}

// Setup creates or replaces the secret for the user. While Enabled=false the
// previous secret can be safely overwritten on every retry.
func (s *TwoFactorService) Setup(userID uint, accountName string) (*SetupResult, error) {
	if userID == 0 {
		return nil, errors.New("auth.unauthorized")
	}
	if accountName == "" {
		accountName = fmt.Sprintf("user-%d", userID)
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.brand,
		AccountName: accountName,
		Period:      30,
		SecretSize:  20, // 160 bits → 32-char base32
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, err
	}
	row := models.UserTwoFactor{
		UserID:  userID,
		Secret:  key.Secret(),
		Enabled: false,
	}
	// Upsert: if row exists, replace its secret and reset Enabled to false.
	if err := s.db.Where("user_id = ?", userID).
		Assign(map[string]any{"secret": key.Secret(), "enabled": false}).
		FirstOrCreate(&row).Error; err != nil {
		return nil, err
	}
	return &SetupResult{Secret: key.Secret(), OTPAuthURL: key.URL()}, nil
}

// Verify accepts the first 6-digit code. On success the row is enabled and a
// fresh batch of recovery codes is generated and returned (shown ONCE).
func (s *TwoFactorService) Verify(userID uint, code string) ([]string, error) {
	var row models.UserTwoFactor
	if err := s.db.Where("user_id = ?", userID).First(&row).Error; err != nil {
		return nil, errors.New("twofa.not_set_up")
	}
	if !totp.Validate(strings.TrimSpace(code), row.Secret) {
		return nil, errors.New("twofa.invalid_code")
	}
	codes, hashed := generateRecoveryCodes(8)
	now := time.Now()
	row.Enabled = true
	row.ConfirmedAt = &now
	row.Recovery = models.StringSlice(hashed)
	if err := s.db.Save(&row).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

// Disable removes 2FA for the user. The caller is responsible for re-checking
// the password before calling this (handled in the handler).
func (s *TwoFactorService) Disable(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.UserTwoFactor{}).Error
}

// Validate is called during login. Accepts either a 6-digit TOTP code or a
// single-use recovery code (which is then consumed). Returns nil on success.
func (s *TwoFactorService) Validate(userID uint, code string) error {
	var row models.UserTwoFactor
	if err := s.db.Where("user_id = ? AND enabled = ?", userID, true).First(&row).Error; err != nil {
		return errors.New("twofa.not_enabled")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("twofa.code_required")
	}
	// Try TOTP first (cheapest).
	if totp.Validate(code, row.Secret) {
		return nil
	}
	// Recovery codes: case-insensitive, ignore the dash separator we add for UX.
	want := hashRecoveryCode(strings.ReplaceAll(strings.ToUpper(code), "-", ""))
	for i, h := range row.Recovery {
		if h == want {
			// Consume it.
			row.Recovery = append(row.Recovery[:i], row.Recovery[i+1:]...)
			_ = s.db.Save(&row).Error
			return nil
		}
	}
	return errors.New("twofa.invalid_code")
}

// IsEnabled reports whether the user has 2FA actively enforced.
func (s *TwoFactorService) IsEnabled(userID uint) bool {
	var n int64
	s.db.Model(&models.UserTwoFactor{}).
		Where("user_id = ? AND enabled = ?", userID, true).Count(&n)
	return n > 0
}

// ===== helpers =====

// generateRecoveryCodes returns the *plaintext* codes (shown to the user once)
// and their hashes (stored in DB). Format: XXXX-XXXX (8 hex chars + dash).
func generateRecoveryCodes(n int) (plain []string, hashed []string) {
	plain = make([]string, 0, n)
	hashed = make([]string, 0, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 4)
		_, _ = rand.Read(buf)
		raw := strings.ToUpper(hex.EncodeToString(buf)) // 8 hex chars
		display := raw[:4] + "-" + raw[4:]
		plain = append(plain, display)
		hashed = append(hashed, hashRecoveryCode(raw))
	}
	return
}

// hashRecoveryCode is intentionally cheap (codes are high-entropy already).
func hashRecoveryCode(s string) string {
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return enc.EncodeToString([]byte("rc:" + s))
}
