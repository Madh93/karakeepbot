package markdown

import (
	"cmp"
	"slices"
	"unicode/utf16"

	"github.com/go-telegram/bot/models"
)

var globalbuf []uint16

func EncodeURLs(s string, entities []models.MessageEntity) string {
	return string(utf16.Decode(encodeURLs(utf16.Encode([]rune(s)), entities)))
}

func encodeURLs(s []uint16, entities []models.MessageEntity) []uint16 {
	entities = slices.Clone(entities)

	slices.SortFunc(
		entities,
		func(e1, e2 models.MessageEntity) int {
			return cmp.Compare(e2.Offset, e1.Offset)
		},
	)

	for _, e := range entities {
		if e.Type == models.MessageEntityTypeTextLink {
			buf := globalbuf[:0]
			buf = append(buf, s[:e.Offset]...)
			buf = append(buf, utf16.Encode([]rune("["))...)
			buf = append(buf, s[e.Offset:e.Offset+e.Length]...)
			buf = append(buf, utf16.Encode([]rune("]("))...)
			buf = append(buf, utf16.Encode([]rune(e.URL))...)
			buf = append(buf, utf16.Encode([]rune(")"))...)
			buf = append(buf, s[e.Offset+e.Length:]...)
			s = append([]uint16(nil), buf...)
		}
	}

	return s
}
