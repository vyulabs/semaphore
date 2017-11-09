package tasks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/db"
)

func (t *task) log(msg string) {
	now := time.Now()

	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":       "log",
			"output":     msg,
			"time":       now,
			"task_id":    t.task.ID,
			"project_id": t.projectID,
		})

		if err != nil {
			panic(err)
		}

		sockets.Message(user, b)
	}

	go func() {
		_, err := db.Mysql.Exec("insert into task__output (task_id, task, output, time) VALUES (?, '', ?, ?)", t.task.ID, msg, now)
		if err != nil {
			fmt.Printf("Failed to insert task output: %s\n", err.Error())
			panic(err)
		}
	}()
}

func (t *task) updateStatus() {
	for _, user := range t.users {
		b, err := json.Marshal(&map[string]interface{}{
			"type":       "update",
			"start":      t.task.Start,
			"end":        t.task.End,
			"status":     t.task.Status,
			"task_id":    t.task.ID,
			"project_id": t.projectID,
		})

		if err != nil {
			panic(err)
		}

		sockets.Message(user, b)
	}

	if _, err := db.Mysql.Exec("update task set status=?, start=?, end=? where id=?", t.task.Status, t.task.Start, t.task.End, t.task.ID); err != nil {
		fmt.Printf("Failed to update task status: %s\n", err.Error())
		t.log("Fatal error with database!")
		panic(err)
	}

	if t.task.Status == "success" {
		_, err := db.Mysql.Exec("update project__template set last_success_task_id=?, " +
			"last_success_build_task_id=?, last_success_version=? where id=?",
				t.task.ID, t.task.BuildTaskID, t.task.Ver, t.task.TemplateID)
		if err != nil {
			fmt.Printf("Failed to update task template: %s\n", err.Error())
			t.log("Fatal error with database!")
			panic(err)
		}
	}
}

func (t *task) logPipe(scanner *bufio.Scanner) {
	for scanner.Scan() {
		t.log(scanner.Text())
	}
}

func (t *task) logCmd(cmd *exec.Cmd) {
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	go t.logPipe(bufio.NewScanner(stderr))
	go t.logPipe(bufio.NewScanner(stdout))
}
