package notif

import (
	"fmt"

	"github.com/bentol/tero/config"
	"gopkg.in/gomail.v2"
)

func SentMailNewUser(username, recipient, stringToken string) error {
	smtpConf := config.Get().SMTP

	SmtpUser := smtpConf.Username
	SmtpPass := smtpConf.Password
	Host := smtpConf.Host
	Port := smtpConf.Port
	Sender := smtpConf.Sender
	SenderName := smtpConf.SenderName
	Subject := "Your teleport account"

	m := gomail.NewMessage()

	body := fmt.Sprintf(
		"Hi %s.\n\n"+
			"You or your lead has requested teleport account for you.\n"+
			"Use this link below to complete the registration.\n"+
			"https://%s/web/newuser/%s"+
			"\n\nThis signup token only valid for 3600 seconds",
		username,
		config.Get().ProxyHost,
		stringToken,
	)

	// Set the alternative part to plain text.
	m.AddAlternative("text/plain", body)

	// Construct the message headers, including a Configuration Set and a Tag.
	m.SetHeaders(map[string][]string{
		"From":    {m.FormatAddress(Sender, SenderName)},
		"To":      {recipient},
		"Subject": {Subject},
	})

	// Send the email.
	d := gomail.NewPlainDialer(Host, Port, SmtpUser, SmtpPass)

	if err := d.DialAndSend(m); err != nil {
		return err
	} else {
		return nil
	}
}
