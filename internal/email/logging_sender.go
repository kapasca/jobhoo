package email

import (
	"context"

	"github.com/jobhoo/jobhoo/internal/database"
)

// LoggingSender wraps an email sender and logs all sends to the database
type LoggingSender struct {
	sender Sender
	logs   *database.EmailLogsRepo
}

// NewLoggingSender creates a new logging-enabled email sender
func NewLoggingSender(sender Sender, logs *database.EmailLogsRepo) *LoggingSender {
	return &LoggingSender{sender: sender, logs: logs}
}

// Send sends an email and logs it to the database
func (ls *LoggingSender) Send(to, subject, htmlBody, textBody string) error {
	// Try to send the email
	err := ls.sender.Send(to, subject, htmlBody, textBody)

	// Determine status and error message
	status := "sent"
	var errorMsg *string
	if err != nil {
		status = "failed"
		errStr := err.Error()
		errorMsg = &errStr
	}

	// Log the email send attempt (ignore logging errors - don't block email send on log failure)
	// Note: We use background context here as we don't want to bind email logging to request context
	_ = ls.logs.Log(context.Background(), nil, to, subject, "email", status, errorMsg)

	return err
}

// SendWithUserID sends an email with user context and logs it
func (ls *LoggingSender) SendWithUserID(ctx context.Context, to, subject, htmlBody, textBody, emailType string, userID *string) error {
	// Try to send the email
	err := ls.sender.Send(to, subject, htmlBody, textBody)

	// Determine status and error message
	status := "sent"
	var errorMsg *string
	if err != nil {
		status = "failed"
		errStr := err.Error()
		errorMsg = &errStr
	}

	// Log the email send attempt
	_ = ls.logs.Log(ctx, userID, to, subject, emailType, status, errorMsg)

	return err
}
