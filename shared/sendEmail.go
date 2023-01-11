package shared

import (
	"encoding/json"
	"eve/service/model"
	"eve/utils"
	"fmt"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

// EmailTask ...
type EmailTask struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

func sendEmail(task model.TaskQueue, gMon *GroupMonitor) (success bool) {
	defer func() {
		success = false
		recover()
		gMon.Done()
	}()

	log := utils.Env.Log
	cfg := utils.Env.Cfg
	if len(cfg.Section("mail").Key("smtp_host").String()) == 0 {
		log.Debug(fmt.Errorf("cant sent mail: email config not available"))
		return false
	}

	mailTask := EmailTask{}
	if err := json.Unmarshal(task.Data, &mailTask); err != nil {
		log.Debug(err)
		return false
	}

	//SMTP Server
	server := mail.NewSMTPClient()
	server.Host = cfg.Section("mail").Key("smtp_host").String()
	server.Port = utils.Atoi(cfg.Section("mail").Key("smtp_port").String())
	server.Username = cfg.Section("mail").Key("smtp_user").String()
	server.Password = cfg.Section("mail").Key("smtp_password").String()

	// encryption
	switch cfg.Section("mail").Key("smtp_encryption").String() {
	case "1":
		server.Encryption = mail.EncryptionSSL
	case "2":
		server.Encryption = mail.EncryptionTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	// authentication
	switch cfg.Section("mail").Key("smtp_auth").String() {
	case "1":
		server.Authentication = mail.AuthLogin
	case "2":
		server.Authentication = mail.AuthCRAMMD5
	default:
		server.Authentication = mail.AuthPlain
	}

	//Variable to keep alive connection
	server.KeepAlive = false
	//Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second
	//Timeout for send the data and wait respond
	server.SendTimeout = 10 * time.Second

	//SMTP client
	smtpClient, err := server.Connect()
	if err != nil {
		log.Debug(err)
		return false
	}

	email := mail.NewMSG()
	email.SetFrom(
		fmt.Sprintf("R.E.A.M <%s>", cfg.Section("mail").Key("smtp_from").String()),
	).
		AddTo(mailTask.To).
		SetSubject(mailTask.Subject)

	email.SetBody(mail.TextHTML, mailTask.HTML)
	email.AddAlternative(mail.TextPlain, mailTask.Text)

	//Call Send and pass the client
	if err = email.Send(smtpClient); err != nil {
		log.Debug(err)
		return false
	}

	log.Debug("Email Sent")

	// msg := mail.NewMessage()
	// msg.SetHeader("To", email.To)
	// msg.SetHeader("Subject", email.Subject)
	// msg.SetHeader("From", cfg.Section("mail").Key("smtp_from").String())
	// msg.SetBody("text/html", email.HTML)
	// msg.SetBody("text/plain", email.Text)

	// if cfg.Section("mail").Key("smtp_security").String() == "1" {
	// 	d.StartTLSPolicy = mail.OpportunisticStartTLS
	// 	log.Debug("smtp_security --> mail.OpportunisticStartTLS")
	// } else if cfg.Section("mail").Key("smtp_security").String() == "2" {
	// 	d.StartTLSPolicy = mail.MandatoryStartTLS
	// 	log.Debug("smtp_security --> mail.MandatoryStartTLS")

	// } else if cfg.Section("mail").Key("smtp_security").String() == "0" {
	// 	d.StartTLSPolicy = mail.NoStartTLS
	// 	log.Debug("smtp_security --> mail.NoStartTLS")
	// }

	// if err := d.DialAndSend(msg); err != nil {
	// 	log.Error(err)
	// 	return false
	// }

	return
}
