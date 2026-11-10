package main

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	plainAmount  = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?$`)
	commaGroups  = regexp.MustCompile(`^[0-9]{1,3}(?:,[0-9]{3})+(?:\.[0-9]+)?$`)
	dotGroups    = regexp.MustCompile(`^[0-9]{1,3}(?:\.[0-9]{3})+,[0-9]+$`)
	commaDecimal = regexp.MustCompile(`^[0-9]+,[0-9]{1,2}$`)
	spaceGroups  = regexp.MustCompile(`^[0-9]{1,3}(?: [0-9]{3})+(?:[.,][0-9]{1,2})?$`)
)

// normalizePrice accepts an amount optionally decorated with its declared
// currency. It validates the entire value; arbitrary text is never stripped to
// digits, which could turn a price range or a negative price into a false price.
func normalizePrice(raw, currency string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "JPY"
	}
	// Currency codes appear before or after the amount in some JSON-LD offers.
	upper := strings.ToUpper(s)
	if strings.HasPrefix(upper, currency) {
		s = strings.TrimSpace(s[len(currency):])
	} else if strings.HasSuffix(upper, currency) {
		s = strings.TrimSpace(s[:len(s)-len(currency)])
	}
	symbols := map[string][]string{"EUR": {"€"}, "JPY": {"¥", "￥"}, "USD": {"US$", "$"}, "GBP": {"£"}, "SEK": {"kr"}}
	for _, symbol := range symbols[currency] {
		if strings.HasPrefix(s, symbol) {
			s = strings.TrimSpace(strings.TrimPrefix(s, symbol))
			break
		}
		if strings.HasSuffix(s, symbol) {
			s = strings.TrimSpace(strings.TrimSuffix(s, symbol))
			break
		}
	}
	// Normalize non-breaking / narrow non-breaking spaces used as group separators.
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
	if spaceGroups.MatchString(s) {
		s = strings.ReplaceAll(s, " ", "")
	}
	switch {
	case plainAmount.MatchString(s):
	case commaGroups.MatchString(s):
		s = strings.ReplaceAll(s, ",", "")
	case dotGroups.MatchString(s):
		s = strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), ",", ".")
	case commaDecimal.MatchString(s):
		s = strings.ReplaceAll(s, ",", ".")
	default:
		return "", errors.New("価格は0以上の数値を入力してください。")
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		return "", errors.New("価格は0以上の数値を入力してください。")
	}
	return s, nil
}
