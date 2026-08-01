package email

// Sender sends emails. Implementations: Dev (logs) and SMTP.
type Sender interface {
	Send(to, subject, htmlBody, textBody string) error
}

// SendWithLogging is used by LoggingSender to send and log emails
type SendWithLogging interface {
	Send(to, subject, htmlBody, textBody string) error
}
