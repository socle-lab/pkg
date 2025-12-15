package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

type PasswordOptions struct {
	Length         int
	MinLower       int
	MinUpper       int
	MinDigits      int
	MinSymbols     int
	AllowAmbiguity bool // ex: 0/O, 1/l/I
}

const (
	lower     = "abcdefghijklmnopqrstuvwxyz"
	upper     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits    = "0123456789"
	symbols   = "!@#$%^&*()-_=+[]{};:,.?/|~"
	ambiguous = "O0Il1"
)

// crypto-rand int in [0, max)
func randInt(max int) (int, error) {
	if max <= 0 {
		return 0, errors.New("max must be > 0")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func pickChar(set string) (byte, error) {
	i, err := randInt(len(set))
	if err != nil {
		return 0, err
	}
	return set[i], nil
}

func shuffleBytes(b []byte) error {
	// Fisher-Yates with crypto/rand
	for i := len(b) - 1; i > 0; i-- {
		j, err := randInt(i + 1)
		if err != nil {
			return err
		}
		b[i], b[j] = b[j], b[i]
	}
	return nil
}

func removeChars(set, toRemove string) string {
	m := make(map[rune]bool, len(toRemove))
	for _, r := range toRemove {
		m[r] = true
	}
	out := make([]rune, 0, len(set))
	for _, r := range set {
		if !m[r] {
			out = append(out, r)
		}
	}
	return string(out)
}

func GenerateStrongPassword(opt PasswordOptions) (string, error) {
	if opt.Length <= 0 {
		opt.Length = 16
	}
	if opt.MinLower == 0 && opt.MinUpper == 0 && opt.MinDigits == 0 && opt.MinSymbols == 0 {
		// defaults "fort"
		opt.MinLower, opt.MinUpper, opt.MinDigits, opt.MinSymbols = 4, 4, 4, 2
	}

	minTotal := opt.MinLower + opt.MinUpper + opt.MinDigits + opt.MinSymbols
	if opt.Length < minTotal {
		return "", fmt.Errorf("length %d is too short; need at least %d", opt.Length, minTotal)
	}

	l := lower
	u := upper
	d := digits
	s := symbols
	if !opt.AllowAmbiguity {
		l = removeChars(l, ambiguous)
		u = removeChars(u, ambiguous)
		d = removeChars(d, ambiguous)
		// symbols: généralement pas besoin
	}

	all := l + u + d + s
	if len(all) == 0 {
		return "", errors.New("no character sets available")
	}

	out := make([]byte, 0, opt.Length)

	// Ensure minimums
	for i := 0; i < opt.MinLower; i++ {
		c, err := pickChar(l)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}
	for i := 0; i < opt.MinUpper; i++ {
		c, err := pickChar(u)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}
	for i := 0; i < opt.MinDigits; i++ {
		c, err := pickChar(d)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}
	for i := 0; i < opt.MinSymbols; i++ {
		c, err := pickChar(s)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}

	// Fill the rest
	for len(out) < opt.Length {
		c, err := pickChar(all)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}

	// Shuffle so required chars aren't predictable at the beginning
	if err := shuffleBytes(out); err != nil {
		return "", err
	}

	return string(out), nil
}
