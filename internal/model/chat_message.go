package model

import "fmt"

type ChatMessage struct {
	ID   string
	Text string
}

func (m ChatMessage) GetText() string {
	return fmt.Sprintf(`{"chat_id": "-%s", "text": "%s\n"}`, m.ID, m.Text)
}
