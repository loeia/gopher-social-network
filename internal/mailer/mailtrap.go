package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"time"

	gomail "gopkg.in/mail.v2"
)

type MailTrapClient struct {
	fromEmail   string
	apiKey      string
	sandboxUser string
	sandboxPass string
}

func NewMailTrapClient(fromEmail, apiKey, user, pass string) (*MailTrapClient, error) {
	if apiKey == "" {
		return nil, errors.New("email api key is required")
	}

	return &MailTrapClient{
		fromEmail:   fromEmail,
		apiKey:      apiKey,
		sandboxUser: user,
		sandboxPass: pass,
	}, nil
}

func (m *MailTrapClient) Send(templateFile string, username string, email string, data any, isSandbox bool) (int, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return -1, err
	}

	message := gomail.NewMessage()

	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())

	message.AddAlternative("text/html", body.String())

	// dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)

	var dialer *gomail.Dialer
	if isSandbox {
		dialer = gomail.NewDialer("sandbox.smtp.mailtrap.io", 2525, m.sandboxUser, m.sandboxPass)
	} else {
		dialer = gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)
	}

	var retryErr error
	for i := range maxRetries {
		if retryErr = dialer.DialAndSend(message); retryErr != nil {
			time.Sleep(time.Duration(1<<i) * time.Second)
			continue
		}
		return 200, nil
	}

	return -1, fmt.Errorf("failed to send email after %d attempt, error: %v", maxRetries, retryErr)
}
