package hx

import (
	"bytes"
	"fmt"
	"unicode/utf16"
)

// jsonStringASCII marshals JSON string values using only ASCII plus \uXXXX escapes
// for non-ASCII runes so HX-Trigger header values stay safe when intermediaries
// mishandle UTF-8 (mojibake like "DÃ©claration").
type jsonStringASCII string

func (s jsonStringASCII) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('"')
	for _, r := range string(s) {
		switch r {
		case '\\':
			buf.WriteString(`\\`)
		case '"':
			buf.WriteString(`\"`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&buf, "\\u%04x", r)
			} else if r < 0x80 {
				buf.WriteRune(r)
			} else if r <= 0xffff {
				fmt.Fprintf(&buf, "\\u%04x", r)
			} else {
				r1, r2 := utf16.EncodeRune(r)
				fmt.Fprintf(&buf, "\\u%04x\\u%04x", r1, r2)
			}
		}
	}
	buf.WriteByte('"')
	return buf.Bytes(), nil
}
