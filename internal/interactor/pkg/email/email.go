package email

import (
	"crypto/tls"
	"hta/internal/interactor/pkg/util/log"

	gomail "gopkg.in/mail.v2"
)

// SendEmail sends email.
func SendEmail(to, fromAddress, fromName, mailPwd, subject, message string) error {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetAddressHeader("From", fromAddress, fromName)

	// Set E-Mail receivers
	m.SetHeader("To", to)

	// Set E-Mail subject
	m.SetHeader("Subject", subject)

	// Set E-Mail body. You can set plain text or html with text/html

	m.SetBody("text/plain", message)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, fromAddress, mailPwd)

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// send E-Mail
	if err := d.DialAndSend(m); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
