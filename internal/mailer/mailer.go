package mailer

import "embed"

const (
	FromName   = "GopherSocialNetwork"
	maxRetries = 3

	UserWelcomeTemplate   = "user_invitation.tmpl"
	PasswordResetTemplate = "password_reset.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(string, string, string, any, bool) (int, error)
}
