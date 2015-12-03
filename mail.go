package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

func sendToMail(fromMail, fromName, password, smtpServer, toMail, toName, subject, body string) error {
	if _, err := net.Dial("tcp", smtpServer); err != nil {
		return err
	}

	hostAddress := strings.Split(smtpServer, ":")[0]
	auth := smtp.PlainAuth(toName, fromMail, password, hostAddress)

	//from := mail.Address{strings.Split(fromMail, "@")[0], fromMail}
	from := mail.Address{fromName, fromMail}
	to := mail.Address{"", toMail}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	err := smtp.SendMail(
		smtpServer,
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	)
	return err
}
