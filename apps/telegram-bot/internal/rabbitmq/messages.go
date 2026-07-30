package rabbitmq

// See packages/schemas/notes-image.json for the JSON Schema definition.
type ImageMessage struct {
	ChatID       int64  `json:"chat_id"`
	MessageID    int    `json:"message_id"`
	FileID       string `json:"file_id"`
	Caption      string `json:"caption,omitempty"`
	MediaGroupID string `json:"media_group_id,omitempty"`
}

// ImageBatchMessage groups multiple images from a single message (album/media group).
type ImageBatchMessage struct {
	BatchID   string   `json:"batch_id"`
	ChatID    int64    `json:"chat_id"`
	MessageID int      `json:"message_id"`
	FileIDs   []string `json:"file_ids"`
	Caption   string   `json:"caption,omitempty"`
}

// NotesResultMessage is published by pipeline-worker after processing is complete.
type NotesResultMessage struct {
	BatchID   string `json:"batch_id"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
	Status    string `json:"status"`
	Text      string `json:"text,omitempty"`
}
