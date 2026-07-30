package bot

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"sqotion/apps/telegram-bot/internal/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	tb "gopkg.in/telebot.v4"
)

type Handler struct {
	Channel *amqp.Channel

	mu      sync.Mutex
	buffers map[string]*imageBuffer // keyed by albumID / message key
}

type imageBuffer struct {
	albumID string
	chatID  int64
	msgID   int
	fileIDs []string
	caption string
	timer   *time.Timer
}

func NewHandler(ch *amqp.Channel) *Handler {
	return &Handler{
		Channel: ch,
		buffers: make(map[string]*imageBuffer),
	}
}

func (h *Handler) Register(b *tb.Bot) {
	b.Handle(tb.OnPhoto, h.onPhoto)
	b.Handle(tb.OnDocument, h.onDocument)
}

func (h *Handler) onPhoto(c tb.Context) error {
	photo := c.Message().Photo
	albumID := c.Message().AlbumID

	c.Bot().React(c.Message().Chat, c.Message(),
		tb.Reactions{
			Reactions: []tb.Reaction{
				{Type: "emoji", Emoji: "✍️"},
			},
			Big: true,
		})

	if albumID == "" {
		// Single image — publish immediately as a single-image batch
		return h.publishSingle(c, photo.FileID)
	}

	// Part of an album — buffer and wait for more
	h.bufferImage(albumID, c.Chat().ID, c.Message().ID, photo.FileID, c.Message().Caption)
	return nil
}

func (h *Handler) onDocument(c tb.Context) error {
	doc := c.Message().Document
	if doc.MIME != "" && isImageMIME(doc.MIME) {
		albumID := c.Message().AlbumID
		if albumID == "" {
			return h.publishSingle(c, doc.FileID)
		}
		h.bufferImage(albumID, c.Chat().ID, c.Message().ID, doc.FileID, c.Message().Caption)
	}
	return nil
}

func (h *Handler) publishSingle(c tb.Context, fileID string) error {
	msg := rabbitmq.ImageBatchMessage{
		BatchID:   newBatchID(),
		ChatID:    c.Chat().ID,
		MessageID: c.Message().ID,
		FileIDs:   []string{fileID},
		Caption:   c.Message().Caption,
	}

	log.Printf("queuing single image chat=%d msg=%d file=%s caption=%q",
		msg.ChatID, msg.MessageID, fileID, msg.Caption)

	return rabbitmq.PublishImageBatch(h.Channel, context.Background(), msg)
}

func (h *Handler) bufferImage(albumID string, chatID int64, msgID int, fileID, caption string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	buf, ok := h.buffers[albumID]
	if !ok {
		batchID := newBatchID()
		buf = &imageBuffer{
			albumID: albumID,
			chatID:  chatID,
			msgID:   msgID,
			caption: caption,
		}
		buf.timer = time.AfterFunc(1*time.Second, func() {
			h.flushBuffer(albumID, batchID)
		})
		h.buffers[albumID] = buf
	}

	buf.fileIDs = append(buf.fileIDs, fileID)
	// Use the first non-empty caption
	if caption != "" && buf.caption == "" {
		buf.caption = caption
	}
}

func (h *Handler) flushBuffer(albumID, batchID string) {
	h.mu.Lock()
	buf, ok := h.buffers[albumID]
	if !ok {
		h.mu.Unlock()
		return
	}
	delete(h.buffers, albumID)
	h.mu.Unlock()

	msg := rabbitmq.ImageBatchMessage{
		BatchID:   batchID,
		ChatID:    buf.chatID,
		MessageID: buf.msgID,
		FileIDs:   buf.fileIDs,
		Caption:   buf.caption,
	}

	log.Printf("queuing album batch chat=%d msg=%d files=%d caption=%q batch=%s",
		msg.ChatID, msg.MessageID, len(msg.FileIDs), msg.Caption, msg.BatchID)

	if err := rabbitmq.PublishImageBatch(h.Channel, context.Background(), msg); err != nil {
		log.Printf("failed to publish album batch: %v", err)
	}
}

func (h *Handler) HandleResult(msg rabbitmq.NotesResultMessage) {
	log.Printf("result received batch=%s chat=%d msg=%d status=%s text=%q",
		msg.BatchID, msg.ChatID, msg.MessageID, msg.Status, msg.Text)
}

func isImageMIME(mime string) bool {
	switch mime {
	case "image/jpeg", "image/png", "image/webp", "image/gif", "image/bmp", "image/tiff":
		return true
	}
	return false
}

func newBatchID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("failed to generate batch ID: %v", err)
	}
	return hex.EncodeToString(b)
}
