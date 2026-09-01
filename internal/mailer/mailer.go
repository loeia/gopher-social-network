package mailer

import "embed"

const (
	FromName   = "GopherSocialNetwork"
	maxRetries = 3

	UserWelcomeTemplate   = "user_invitation.tmpl"
	PasswordResetTemplate = "password_reset.tmpl"
	AccountBanTemplate    = "account_ban.tmpl"
	AccountUnbanTemplate  = "account_unban.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(string, string, any, bool) (int, error)
}
