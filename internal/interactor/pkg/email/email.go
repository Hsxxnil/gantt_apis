package email

import (
	"crypto/tls"
	"hta/config"
	"hta/internal/interactor/pkg/util/log"

	gomail "gopkg.in/mail.v2"
)

// SendEmailWithText sends email with text.
func SendEmailWithText(to, fromName, subject, message string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", config.MailAddress, fromName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, config.MailAddress, config.MailPassword)

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

// SendEmailWithHtml sends email with html.
func SendEmailWithHtml(to, fromName, subject, message string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", config.MailAddress, fromName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, config.MailAddress, config.MailPassword)

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
