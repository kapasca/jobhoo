package email

const (
	TypeEmailVerification = "email_verification"
	TypePasswordReset     = "password_reset"
	TypeApplicationStatus = "application_status"
	TypeNewApplication    = "new_application"
	TypeJobClosed         = "job_closed"
	TypeJobReminder       = "job_reminder"
	TypeCompanyApproval   = "company_approval"
	TypeNotification      = "notification"
)

// EmailContent holds the content for an email
type EmailContent struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}
