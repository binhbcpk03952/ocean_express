package utils

import (
	"strings"
	"unicode/utf8"
)

// cp1252RuneToByte ánh xạ các ký tự Unicode sinh ra từ Windows-1252 (0x80 - 0x9F) về lại byte gốc
var cp1252RuneToByte = map[rune]byte{
	'\u20AC': 0x80, // €
	'\u201A': 0x82, // ‚
	'\u0192': 0x83, // ƒ
	'\u201E': 0x84, // „
	'\u2026': 0x85, // …
	'\u2020': 0x86, // †
	'\u2021': 0x87, // ‡
	'\u02C6': 0x88, // ˆ
	'\u2030': 0x89, // ‰
	'\u0160': 0x8A, // Š
	'\u2039': 0x8B, // ‹
	'\u0152': 0x8C, // Œ
	'\u017D': 0x8E, // Ž
	'\u2018': 0x91, // ‘
	'\u2019': 0x92, // ’
	'\u201C': 0x93, // “
	'\u201D': 0x94, // ”
	'\u2022': 0x95, // •
	'\u2013': 0x96, // –
	'\u2014': 0x97, // —
	'\u02DC': 0x98, // ˜
	'\u2122': 0x99, // ™
	'\u0161': 0x9A, // š
	'\u203A': 0x9B, // ›
	'\u0153': 0x9C, // œ
	'\u017E': 0x9E, // ž
	'\u0178': 0x9F, // Ÿ
}

func isLatin1OrCP1252(r rune) bool {
	if r <= 0xFF {
		return true
	}
	_, ok := cp1252RuneToByte[r]
	return ok
}

// FixMojibake phát hiện và tự động giải mã chuỗi UTF-8 bị double-encode / mojibake
// (thường xảy ra khi chuỗi UTF-8 tiếng Việt bị decode nhầm bằng ISO-8859-1 hoặc Windows-1252 rồi re-encode UTF-8).
func FixMojibake(s string) string {
	if s == "" {
		return ""
	}

	// 1. Thử giải mã toàn bộ chuỗi trước
	if fixed, ok := tryDecodeChunk(s); ok && fixed != s {
		return fixed
	}

	// 2. Thử tách theo dấu phẩy (các phân đoạn địa chỉ như "300/6 HÃ Huy Táº-p, Phường Tân An")
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		changed := false
		for i, p := range parts {
			if fixedPart, ok := tryDecodeChunk(p); ok && fixedPart != p {
				parts[i] = fixedPart
				changed = true
			} else if fixedPartSub := fixSubRuns(p); fixedPartSub != p {
				parts[i] = fixedPartSub
				changed = true
			}
		}
		if changed {
			return strings.Join(parts, ",")
		}
	}

	return fixSubRuns(s)
}

// fixSubRuns quét và giải mã các đoạn rune Latin-1/CP1252 liên tiếp nằm xen kẽ với rune Unicode mở rộng
func fixSubRuns(s string) string {
	var builder strings.Builder
	var currentChunk []rune

	flush := func() {
		if len(currentChunk) == 0 {
			return
		}
		chunkStr := string(currentChunk)
		if decoded, ok := tryDecodeChunk(chunkStr); ok && decoded != chunkStr {
			builder.WriteString(decoded)
		} else {
			builder.WriteString(chunkStr)
		}
		currentChunk = currentChunk[:0]
	}

	for _, r := range s {
		if isLatin1OrCP1252(r) {
			currentChunk = append(currentChunk, r)
		} else {
			flush()
			builder.WriteRune(r)
		}
	}
	flush()

	return builder.String()
}

// tryDecodeChunk thử chuyển đổi các rune trong chunk về byte và kiểm tra tính hợp lệ UTF-8
func tryDecodeChunk(s string) (string, bool) {
	if s == "" {
		return s, false
	}

	var rawBytes []byte
	hasHighByte := false

	for _, r := range s {
		if b, ok := cp1252RuneToByte[r]; ok {
			rawBytes = append(rawBytes, b)
			hasHighByte = true
		} else if r <= 0xFF {
			rawBytes = append(rawBytes, byte(r))
			if r >= 0x80 {
				hasHighByte = true
			}
		} else {
			// Chứa rune > 0xFF không thuộc CP1252
			return s, false
		}
	}

	if !hasHighByte {
		return s, false
	}

	if utf8.Valid(rawBytes) {
		decoded := string(rawBytes)
		// Chỉ chấp nhận nếu sau khi decode số rune giảm (chứng minh các byte ghép lại thành ký tự UTF-8 đa byte hợp lệ)
		if utf8.RuneCountInString(decoded) < utf8.RuneCountInString(s) {
			return decoded, true
		}
	}

	return s, false
}
