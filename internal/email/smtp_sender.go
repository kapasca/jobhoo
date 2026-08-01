package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{host: host, port: port, username: username, password: password, from: from}
}

func (s *SMTPSender) Send(to, subject, htmlBody, textBody string) error {
	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	header := make(map[string]string)
	header["From"] = s.from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "multipart/alternative; boundary=BOUNDARY"

	var msg strings.Builder
	for k, v := range header {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n--BOUNDARY\r\n")
	msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n--BOUNDARY\r\n")
	msg.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n--BOUNDARY--")

	// Dial plain connection and then upgrade to TLS with STARTTLS
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Quit()

	// Upgrade to TLS
	tlsconfig := &tls.Config{InsecureSkipVerify: true, ServerName: s.host}
	if err = c.StartTLS(tlsconfig); err != nil {
		return err
	}

	if err = c.Auth(auth); err != nil {
		return err
	}
	if err = c.Mail(s.from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg.String()))
	if err != nil {
		return err
	}
	err = w.Close()
	return err
}
