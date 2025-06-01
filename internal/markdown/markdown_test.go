package markdown_test

import (
	"testing"

	"github.com/Madh93/karakeepbot/internal/markdown"
	"github.com/go-telegram/bot/models"
)

func TestEncodeURLs(t *testing.T) {
	f := func(t *testing.T, name string, s string, entities []models.MessageEntity, want string) {
		t.Helper()

		t.Run(name, func(t *testing.T) {
			t.Helper()

			if got := markdown.EncodeURLs(s, entities); got != want {
				t.Errorf("unexpected: %s", got)
			}
		})
	}

	mkurl := func(offset, length int, url string) models.MessageEntity {
		return models.MessageEntity{
			Type:   models.MessageEntityTypeTextLink,
			Offset: offset,
			Length: length,
			URL:    url,
		}
	}

	f(
		t,
		"one url",
		"foo bar baz",
		[]models.MessageEntity{mkurl(4, 3, "http://yahoo.com")},
		"foo [bar](http://yahoo.com) baz",
	)

	f(
		t,
		"two urls",
		"foo bar baz",
		[]models.MessageEntity{
			mkurl(0, 3, "http://google.com"),
			mkurl(4, 3, "http://yahoo.com"),
		},
		"[foo](http://google.com) [bar](http://yahoo.com) baz",
	)
}
