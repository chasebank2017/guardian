package handler

type UploadMessagesPayload struct {
	Messages []*ChatMessagePayload `json:"messages" validate:"required,dive"`
}

type ChatMessagePayload struct {
	Content   string `json:"content" validate:"required,max=10000"`
	Timestamp int64  `json:"timestamp" validate:"required"`
}
