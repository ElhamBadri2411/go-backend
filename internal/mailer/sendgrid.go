package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendgrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}
	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	// isSandbox = false
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	for i := 0; i < MaxRetries; i++ {
		res, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %v, attempt %d / %d", email, i, MaxRetries)
			log.Printf("Error: %v", err.Error())

			// exponential backoff
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		} else {
			log.Printf("SendGrid response (attempt %d): status=%d, body=%q",
				i+1, res.StatusCode, res.Body)
		}

		log.Printf("Email sent successfully to %v", email)
		return nil
	}

	return fmt.Errorf("Failed to send email to %v after %d tries", email, MaxRetries)
}
