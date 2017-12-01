package tasks

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"strings"
)

const emailTemplate = `Subject: Task '{{ .Alias }}' failed

Task {{ .TaskID }} with template '{{ .Alias }}' has failed!
Task log: <a href='{{ .TaskURL }}'>{{ .TaskURL }}</a>`

const telegramTemplate = `{"chat_id": "{{ .ChatID }}","parse_mode":"HTML","text":"<code>{{ .Alias }}</code>\n#{{ .TaskID }} <b>{{ .TaskResult }}</b> {{ .TaskVersion }} {{ .TaskDescription }}\nby{{ .Author }}"}`
type Alert struct {
	TaskID  string
	Alias   string
	TaskURL string
	ChatID  string
	TaskResult string
	TaskColor string
	TaskDescription string
	TaskVersion string
	Author string
}

func (t *task) sendMailAlert() {
	if util.Config.EmailAlert != true || t.alert != true {
		return
	}

	mailHost := util.Config.EmailHost + ":" + util.Config.EmailPort

	var mailBuffer bytes.Buffer
	alert := Alert{
		TaskID:  strconv.Itoa(t.task.ID),
		Alias:   t.template.Alias,
		TaskURL: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID),
	}
	tpl := template.New("mail body template")
	tpl, err := tpl.Parse(emailTemplate)
	err = tpl.Execute(&mailBuffer, alert)

	if err != nil {
		t.log("Can't generate alert template!")
		panic(err)
	}

	for _, user := range t.users {
		userObj, err := db.FetchUser(user)

		if userObj.Alert != true {
			return
		}

		if err != nil {
			t.log("Can't find user Email!")
			panic(err)
		}

		t.log("Sending email to " + userObj.Email + " from " + util.Config.EmailSender)
		err = util.SendMail(mailHost, util.Config.EmailSender, userObj.Email, mailBuffer)
		if err != nil {
			t.log("Can't send email!")
			t.log("Error: " + err.Error())
			panic(err)
		}
	}
}

func (t *task) sendTelegramAlert() {
	if util.Config.TelegramAlert != true || t.alert != true {
		return
	}

	chat_id := util.Config.TelegramChat
	if t.alert_chat != "" {
		chat_id = t.alert_chat
	}

	var telegramBuffer bytes.Buffer

	var ver string
	if t.task.Ver != nil {
		ver = *t.task.Ver
	} else if t.task.BuildTaskID != nil {
		ver = "build " + strconv.Itoa(*t.task.BuildTaskID)
	} else {
		ver = ""
	}


	var description string
	if t.task.Description != nil {
		description = "- " + *t.task.Description
	}

	var author string
	var user db.User
	if err := db.Mysql.SelectOne(&user, "select * from user where id=?", t.task.UserID); err == nil {
		author = user.Name
	}

	alert := Alert{
		TaskID:  strconv.Itoa(t.task.ID),
		Alias:   t.template.Alias,
		TaskURL: util.Config.WebHost + "/project/" + strconv.Itoa(t.template.ProjectID),
		ChatID:  chat_id,
		TaskResult: strings.ToUpper(t.task.Status),
		TaskColor: "red",
		TaskVersion: ver,
		TaskDescription: description,
		Author: author,
	}


	switch t.task.Status {
	case "success":
		alert.TaskColor = "green"
	}

	tpl := template.New("telegram body template")
	tpl, err := tpl.Parse(telegramTemplate)
	err = tpl.Execute(&telegramBuffer, alert)

	if err != nil {
		t.log("Can't generate alert template!")
		panic(err)
	}

	resp, err := http.Post("https://api.telegram.org/bot"+util.Config.TelegramToken+"/sendMessage", "application/json", &telegramBuffer)

	if err != nil {
		t.log("Can't send telegram alert!")
		panic(err)
	}

	if resp.StatusCode != 200 {
		t.log("Can't send telegram alert! Response code not 200!")
	}
}
