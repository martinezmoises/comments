package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"time"

	"github.com/go-mail/mail/v2"
)

// When we build our app for production, we want to include the templates
// so we will use the go emded feature which creates a virtual file system
// for us. To do this, we will use a special syntax that looks like a comment
// but it is not a comment '//go:embed "templates"', it is a directive to Go
// to build a virtual filesystem for the 'templates directory'

// this allow us to not need a  separate server for serving static files
//
//go:embed "templates"
var templateFS embed.FS // embed the files from templates into our program

// dialer is a connection to the SMTP server
// sender is who is sending the email to the new user
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// Configure a SMTP connection instance using our credentials from Mailtrap
func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send the email to the user. The data parameter is for the dynamic data
// to inject into the template
func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		log.Printf("Error parsing email template: %v", err)
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		log.Printf("Error executing subject template: %v", err)
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		log.Printf("Error executing plain body template: %v", err)
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		log.Printf("Error executing HTML body template: %v", err)
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			log.Printf("Email sent successfully to %s", recipient)
			return nil
		}
		log.Printf("Attempt %d to send email to %s failed: %v", i, recipient, err)
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Failed to send email to %s after 3 attempts", recipient)
	return err
}
