package utils

import (
	"encoding/base64"
	"fmt"
	smtpNet "net/smtp"
	"strings"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"github.com/spf13/viper"
)

type SendMailModel struct {
	SenderName   string `json:"sender_name"`
	SenderMail   string `json:"sender_mail"`
	ReceiverName string `json:"receiver_name"`
	ReceiverMail string `json:"receiver_mail"`
	Subject      string `json:"subject"`
	Body         string `json:"body"`
	AltBody      string `json:"alt_body"`

	Username string `json:"username"`
	Password string `json:"password"`
}

func SendMail(m SendMailModel) (bool, error) {
	if m.ReceiverMail == "" {
		return false, nil
	}
	if m.Body == "" {
		m.Body = m.AltBody
	}

	smtpHost := "smtp.kku.ac.th"
	smtpPort := 25

	from := mail.Address{Name: m.SenderName, Address: m.SenderMail}
	to := mail.Address{Name: m.ReceiverName, Address: m.ReceiverMail}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = m.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	var message strings.Builder
	for k, v := range header {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(base64.StdEncoding.EncodeToString([]byte(m.Body)))

	client, err := smtp.Dial(fmt.Sprintf("%s:%d", smtpHost, smtpPort))
	if err != nil {
		return false, err
	}
	defer client.Close()

	if err = client.Mail(from.Address, nil); err != nil {
		return false, err
	}
	if err = client.Rcpt(to.Address, nil); err != nil {
		return false, err
	}
	wc, err := client.Data()
	if err != nil {
		return false, err
	}
	defer wc.Close()

	_, err = wc.Write([]byte(message.String()))
	if err != nil {
		return false, err
	}

	client.Quit()

	return true, nil
}
func SendNormalMail(m SendMailModel) (bool, error) {
	if m.ReceiverMail == "" {
		return false, nil
	}
	if m.Body == "" {
		m.Body = m.AltBody
	}

	smtpHost := viper.GetString("app.smtp.host_test")
	smtpPort := viper.GetInt("app.smtp.port_test")
	m.Username = viper.GetString("app.smtp.username_test")
	m.Password = viper.GetString("app.smtp.password_test")
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

	from := mail.Address{Name: m.SenderName, Address: m.SenderMail}
	to := mail.Address{Name: m.ReceiverName, Address: m.ReceiverMail}

	// Build message
	header := []string{
		"From: " + from.String(),
		"To: " + to.String(),
		"Subject: " + m.Subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"utf-8\"",
		"Content-Transfer-Encoding: base64",
	}
	var msg strings.Builder
	for _, h := range header {
		msg.WriteString(h + "\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(base64.StdEncoding.EncodeToString([]byte(m.Body)))

	// AUTH
	auth := smtpNet.PlainAuth("", m.Username, m.Password, smtpHost)

	// Send
	if err := smtpNet.SendMail(addr, auth, from.Address, []string{to.Address}, []byte(msg.String())); err != nil {
		return false, err
	}
	return true, nil
}
