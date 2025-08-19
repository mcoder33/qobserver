package slimtg

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Message interface {
	GetText() string
}

type ChatMessage struct {
	ID   string
	Text string
}

func (m ChatMessage) GetText() string {
	return fmt.Sprintf(`{"chat_id": "-%s", "text": "%s\n"}`, m.ID, m.Text)
}

type Client struct {
	token   string
	verbose bool
}

func NewClient(token string) *Client {
	return &Client{token: token, verbose: false}
}

func (c *Client) VerboseMode() {
	c.verbose = true
}

func (c *Client) Send(message Message) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)
	buf := strings.NewReader(message.GetText())

	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return fmt.Errorf("ERROR: failed to send message: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("WARN: failed to close response body: %v", err)
		}
	}()

	if c.verbose {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("ERROR: failed to read response body: %w", err)
		}
		log.Printf("INFO: Response: %s, body: %s", resp.Status, string(b))
	}

	return nil
}
