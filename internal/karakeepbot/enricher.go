package karakeepbot

import "context"

// enrichBookmark adds Telegram origin metadata to a newly created bookmark.
// Currently attaches the #telegram tag. Non-fatal on failure.
func (kb *KarakeepBot) enrichBookmark(ctx context.Context, msg TelegramMessage, bookmark *KarakeepBookmark) {
	if err := kb.karakeep.AddTag(ctx, bookmark.Id, "telegram"); err != nil {
		kb.logger.Warn("Failed to add #telegram tag", "bookmark_id", bookmark.Id, "error", err)
	}
}
