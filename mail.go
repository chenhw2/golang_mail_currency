package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"regexp"
	"strings"
)

var isSTMP = regexp.MustCompile(`^(\w+\.)+(\w+):\d+$`)

func sendToMail(fromMail, fromName, password, smtpServer string, toMails []string, subject, body string) (err error) {
	if _, err := net.Dial("tcp", smtpServer); err != nil {
		return err
	}

	if !isSTMP.MatchString(smtpServer) {
		return errors.New("Smtp Server Check Fail")
	}

	from, to, message := mail.Address{Name: fromName, Address: fromMail}, "", ""
	for _, mail := range toMails {
		to += mail + ";"
	}
	to = to[:len(to)-1] // remove last ';'
	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to
	header["Subject"] = "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = `text/plain; charset="utf-8"`
	header["Content-Transfer-Encoding"] = "base64"

	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	err = smtp.SendMail(
		smtpServer,
		smtp.PlainAuth("", fromMail, password, strings.Split(smtpServer, ":")[0]),
		from.Address,
		toMails,
		[]byte(message),
	)
	return
}
