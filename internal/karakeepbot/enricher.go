package karakeepbot

import "context"

// enrichBookmark adds Telegram origin metadata to a newly created bookmark.
// It attaches the #telegram tag plus any hashtags found in the message text or
// photo caption. Non-fatal on failure.
func (kb *KarakeepBot) enrichBookmark(ctx context.Context, msg TelegramMessage, bookmark *KarakeepBookmark) {
	tags := []string{"telegram"}
	for _, tag := range msg.Hashtags() {
		if tag != "telegram" {
			tags = append(tags, tag)
		}
	}

	if err := kb.karakeep.AddTags(ctx, bookmark.Id, tags); err != nil {
		kb.logger.Warn("Failed to add tags", "bookmark_id", bookmark.Id, "tags", tags, "error", err)
	}
}
