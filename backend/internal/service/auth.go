package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/instaagrammeta/somon-crm/backend/internal/config"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Claims is the JWT payload.
type Claims struct {
	UserID   uint   `json:"uid"`
	Login    string `json:"login"`
	FullName string `json:"name"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	cfg config.JWTConfig
	db  *gorm.DB
	rdb *redis.Client
}

func NewAuth(cfg config.JWTConfig, db *gorm.DB, rdb *redis.Client) *AuthService {
	return &AuthService{cfg: cfg, db: db, rdb: rdb}
}

func (s *AuthService) RedisTimeout() time.Duration { return 2 * time.Second }

// HashPassword wraps bcrypt with a sensible cost.
func (s *AuthService) HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *AuthService) CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// IssueToken creates a signed JWT carrying user identity.
func (s *AuthService) IssueToken(u *models.User) (string, time.Time, error) {
	exp := time.Now().Add(s.cfg.AccessTTL)
	claims := Claims{
		UserID:   u.ID,
		Login:    u.Login,
		FullName: u.FullName,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   u.Login,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.Secret))
	return signed, exp, err
}

// ParseToken validates signature & expiry; returns claims.
func (s *AuthService) ParseToken(raw string) (*Claims, error) {
	c := &Claims{}
	token, err := jwt.ParseWithClaims(raw, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token invalid")
	}
	return c, nil
}

// Blacklist marks the token's JTI as revoked until its natural expiry.
func (s *AuthService) Blacklist(ctx context.Context, claims *Claims) error {
	if s.rdb == nil {
		return nil
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, "jwt:bl:"+claims.ID, "1", ttl).Err()
}

func (s *AuthService) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if s.rdb == nil {
		return false, nil
	}
	v, err := s.rdb.Get(ctx, "jwt:bl:"+jti).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return v == "1", nil
}
