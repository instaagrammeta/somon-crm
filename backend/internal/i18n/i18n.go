// Package i18n provides minimal translation lookup for backend error / notification messages.
// Two locales are supported: "tg" (Tajik, default) and "ru" (Russian).
package i18n

import "strings"

const (
	LocaleTG = "tg"
	LocaleRU = "ru"
)

// Translate returns the message for given key in the resolved locale.
// Falls back to Tajik, then English-style key, if not found.
func Translate(locale, key string, args ...any) string {
	locale = NormalizeLocale(locale)
	bundle, ok := bundles[locale]
	if !ok {
		bundle = bundles[LocaleTG]
	}
	if msg, ok := bundle[key]; ok {
		return formatArgs(msg, args...)
	}
	if msg, ok := bundles[LocaleTG][key]; ok {
		return formatArgs(msg, args...)
	}
	return key
}

// NormalizeLocale extracts a known locale from a raw header value.
func NormalizeLocale(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return LocaleTG
	}
	// Accept-Language can be "ru-RU,ru;q=0.9,tg;q=0.8" — pick first known
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		code := strings.TrimSpace(strings.Split(p, ";")[0])
		code = strings.SplitN(code, "-", 2)[0]
		if code == LocaleRU || code == LocaleTG {
			return code
		}
	}
	return LocaleTG
}

func formatArgs(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	// Simple positional replace: {0}, {1}, ...
	out := msg
	for i, a := range args {
		placeholder := "{" + itoa(i) + "}"
		out = strings.ReplaceAll(out, placeholder, toString(a))
	}
	return out
}

func itoa(i int) string {
	switch i {
	case 0:
		return "0"
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	}
	// Fallback general case
	if i < 0 {
		return "-" + itoa(-i)
	}
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + itoa(i%10)
}

func toString(a any) string {
	switch v := a.(type) {
	case string:
		return v
	case int:
		return itoa(v)
	default:
		return ""
	}
}
