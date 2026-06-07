package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

// MakeSlug returns a URL-safe transliterated slug.
func MakeSlug(s string) string {
	out := slug.Make(s)
	if out == "" {
		out = uuid.New().String()[:8]
	}
	return out
}

// EnsureUniqueSlug appends -2, -3, ... if a "is unique" check fails.
func EnsureUniqueSlug(base string, isFree func(s string) bool) string {
	if isFree(base) {
		return base
	}
	for i := 2; i < 1000; i++ {
		cand := fmt.Sprintf("%s-%d", base, i)
		if isFree(cand) {
			return cand
		}
	}
	return fmt.Sprintf("%s-%s", base, uuid.New().String()[:6])
}
