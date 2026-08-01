package email

import (
	"fmt"
	"log"
)

type DevSender struct{}

func NewDevSender() *DevSender { return &DevSender{} }

func (s *DevSender) Send(to, subject, htmlBody, textBody string) error {
	// In development we log the message so developers can copy link/token.
	log.Printf("[DEV EMAIL] To:%s Subject:%s\nText:\n%s\nHTML:\n%s\n", to, subject, textBody, htmlBody)
	// Also print a short console-friendly one-line to help quick copy.
	fmt.Printf("DEV EMAIL -> %s | %s\n", to, subject)
	return nil
}
