package tasks

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
	"database/sql"
	"strings"
	"bytes"
	"errors"
)

func resolveDefaultVersion(versionTemplate string, taskID int, taskNum int) string {
	ret := strings.Replace(versionTemplate, "<next_index>", "0", -1)
	ret = strings.Replace(ret, "<task_id>", strconv.Itoa(taskID), -1)
	ret = strings.Replace(ret, "<task_num>", strconv.Itoa(taskNum), -1)
	return ret
}

func ResolveNewVersion(currentVersion string, versionTemplate string, taskID int, taskNum int) (string, error) {
	if currentVersion  == "" {
		return resolveDefaultVersion(versionTemplate, taskID, taskNum), nil
	}

	const text = 0
	const field = 1

	var ret bytes.Buffer
	var fieldName bytes.Buffer

	state := text

	k := 0

	for i := 0; i < len(versionTemplate); i++ {
		switch c := versionTemplate[i]; c {
		case '<':
			state = field
		case '>':
			state = text
			switch strings.Trim(fieldName.String(), " ") {
			case "next_index":
				var end int
				if i + 1 < len(versionTemplate) {
					for v:=k; v < len(currentVersion); v++ {
						if versionTemplate[i + 1] == currentVersion[v] {
							end = v
							break
						}
					}
					if end == len(currentVersion) {
						return "", errors.New("illegal version format")
					}
				} else {
					end = len(currentVersion)
				}
				current, err := strconv.Atoi(currentVersion[k:end])
				k = end
				if err != nil {
					return "", err
				}
				ret.WriteString(strconv.Itoa(current + 1))
			case "task_id":
				ret.WriteString(strconv.Itoa(taskID))
			case "task_num":
				ret.WriteString(strconv.Itoa(taskNum))
			}
			fieldName.Reset()
			state = text
		default:
			if state == text {
				if versionTemplate[i] == currentVersion[k] {
					ret.WriteByte(versionTemplate[i])
					k++
				} else {
					return resolveDefaultVersion(versionTemplate, taskID, taskNum), nil
				}
			} else if state == field {
				fieldName.WriteByte(versionTemplate[i])
			}
		}
	}

	return ret.String(), nil
}

func AddTask(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	user := context.Get(r, "user").(*db.User)

	var taskObj db.Task
	if err := mulekick.Bind(w, r, &taskObj); err != nil {
		return
	}

	taskObj.Created = time.Now()
	taskObj.Status = "waiting"
	taskObj.UserID = &user.ID

	if taskObj.BuildTaskID != nil {
		var buildObj db.Task
		if err := db.Mysql.SelectOne(&buildObj, "select * from task where id=?", taskObj.BuildTaskID); err != nil {
			panic(err)
		}
		taskObj.Commit = buildObj.Commit
	}

	if err := db.Mysql.Insert(&taskObj); err != nil {
		panic(err)
	}

	const prevTaskNumQuery = "select * from task where id<? order by id desc limit 1"
	var prevTaskObj db.Task

	for sum := 0; sum < 4; sum++ {
		if err := db.Mysql.SelectOne(&prevTaskObj, prevTaskNumQuery, taskObj.ID); err != nil {
			if err == sql.ErrNoRows {
				num := 0
				prevTaskObj.Num = &num
			}
			panic(err)
		}
		if prevTaskObj.Num == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	if prevTaskObj.Num == nil {
		num := 0
		prevTaskObj.Num = &num
		//panic("Can't resolve number of previous task")
	}

	taskNum := *prevTaskObj.Num + 1
	taskObj.Num = &taskNum

	var templateObj db.Template
	if err := db.Mysql.SelectOne(&templateObj, "select * from project__template where id=?", taskObj.TemplateID); err != nil {
		panic(err)
	}

	if templateObj.VersionTemplate != nil {
		//version = strings.Replace(*templateObj.VersionTemplate, "{{ task_id }}", strconv.Itoa(taskObj.ID), -1)
		//version = strings.Replace(version, "{{ task_num }}", strconv.Itoa(taskNum), -1)
		var prevVer string
		if prevTaskObj.Ver != nil {
			prevVer = *prevTaskObj.Ver
		} else {
			prevVer = ""
		}
		if version, err := ResolveNewVersion(prevVer, *templateObj.VersionTemplate, taskObj.ID, taskNum); err != nil {
			taskObj.Ver = &version
		} else {
			println(err)
			panic(err)
		}
	}

	if _, err := db.Mysql.Update(&taskObj); err != nil {
		panic(err)
	}

	pool.register <- &task{
		task:      taskObj,
		projectID: project.ID,
	}

	objType := "task"
	desc := "Task ID " + strconv.Itoa(taskObj.ID) + " queued for running"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &taskObj.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, taskObj)
}

func GetTasksList(w http.ResponseWriter, r *http.Request, limit uint64) {
	project := context.Get(r, "project").(db.Project)

	q := squirrel.Select("task.*, tpl.playbook as tpl_playbook, user.name as user_name, tpl.alias as tpl_alias").
		From("task").
		Join("project__template as tpl on task.template_id=tpl.id").
		LeftJoin("user on task.user_id=user.id");

	if tpl := context.Get(r, "template"); tpl != nil {
		q = q.Where("tpl.project_id=? AND task.template_id=?", project.ID, tpl.(db.Template).ID)
	} else {
		q = q.Where("tpl.project_id=?", project.ID)
	}

	q = q.OrderBy("task.created desc, id desc")

	if limit > 0 {
		q = q.Limit(limit)
	}

	query, args, _ := q.ToSql()

	var tasks []struct {
		db.Task

		TemplatePlaybook string  `db:"tpl_playbook" json:"tpl_playbook"`
		TemplateAlias    string  `db:"tpl_alias" json:"tpl_alias"`
		UserName         *string `db:"user_name" json:"user_name"`
	}
	if _, err := db.Mysql.Select(&tasks, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, tasks)
}

func GetAllTasks(w http.ResponseWriter, r *http.Request) {
	GetTasksList(w, r, 0)
}

func GetLastTasks(w http.ResponseWriter, r *http.Request) {
	GetTasksList(w, r, 10)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)
	mulekick.WriteJSON(w, http.StatusOK, task)
}

func GetTaskMiddleware(w http.ResponseWriter, r *http.Request) {
	taskID, err := util.GetIntParam("task_id", w, r)
	if err != nil {
		panic(err)
	}

	var task db.Task
	if err := db.Mysql.SelectOne(&task, "select * from task where id=?", taskID); err != nil {
		panic(err)
	}

	context.Set(r, "task", task)
}

func GetTaskOutput(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)

	var output []db.TaskOutput
	if _, err := db.Mysql.Select(&output, "select task_id, task, time, output from task__output where task_id=? order by time asc", task.ID); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, output)
}

func RemoveTask(w http.ResponseWriter, r *http.Request) {
	task := context.Get(r, "task").(db.Task)

	statements := []string{
		"delete from task__output where task_id=?",
		"delete from task where id=?",
	}

	for _, statement := range statements {
		_, err := db.Mysql.Exec(statement, task.ID)
		if err != nil {
			panic(err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
