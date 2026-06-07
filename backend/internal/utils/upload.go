package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var allowedExt = map[string]string{
	".png": "image", ".jpg": "image", ".jpeg": "image", ".gif": "image",
	".webp": "image", ".bmp": "image", ".svg": "image",
	".mp4": "video", ".webm": "video", ".mov": "video", ".avi": "video", ".mkv": "video", ".flv": "video",
	".pdf":  "pdf",
	".xls":  "document", ".xlsx": "document", ".doc": "document", ".docx": "document",
	".csv":  "document", ".txt": "document",
	".zip":  "archive", ".rar": "archive", ".7z": "archive",
}

var ErrUnsupportedExt = errors.New("file extension not allowed")

// SafeFilename returns a sanitized filename with timestamp + random hex prefix.
func SafeFilename(orig string) string {
	orig = filepath.Base(orig)
	ext := strings.ToLower(filepath.Ext(orig))
	stem := strings.TrimSuffix(orig, ext)
	stem = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, stem)
	if stem == "" {
		stem = "file"
	}
	rnd := make([]byte, 4)
	_, _ = rand.Read(rnd)
	return time.Now().UTC().Format("20060102_150405") + "_" + hex.EncodeToString(rnd) + "_" + stem + ext
}

// SaveUpload writes the multipart file under uploadDir/subdir, returns relative URL path.
// fileType is one of: image, video, pdf, document, archive, "" (other).
func SaveUpload(file *multipart.FileHeader, uploadDir, subdir string) (relPath, fileType string, err error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	t, ok := allowedExt[ext]
	if !ok {
		return "", "", ErrUnsupportedExt
	}
	fileType = t

	if err := os.MkdirAll(filepath.Join(uploadDir, subdir), 0o755); err != nil {
		return "", "", err
	}
	name := SafeFilename(file.Filename)
	full := filepath.Join(uploadDir, subdir, name)

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()
	dst, err := os.Create(full)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", "", err
	}

	rel := "/uploads/" + filepath.ToSlash(filepath.Join(subdir, name))
	return rel, fileType, nil
}
