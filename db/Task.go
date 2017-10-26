package db

import "time"

type Task struct {
	ID         int `db:"id" json:"id"`
	TemplateID int `db:"template_id" json:"template_id" binding:"required"`

	Status string `db:"status" json:"status"`
	Debug  bool   `db:"debug" json:"debug"`

	DryRun bool `db:"dry_run" json:"dry_run"`

	// override variables
	Playbook    string `db:"playbook" json:"playbook"`
	Environment string `db:"environment" json:"environment"`
	Description *string `db:"description" json:"description"`
	BuildTaskID *int `db:"build_task_id" json:"build_task_id"`

	UserID *int `db:"user_id" json:"user_id"`

	Created time.Time  `db:"created" json:"created"`
	Start   *time.Time `db:"start" json:"start"`
	End     *time.Time `db:"end" json:"end"`
}

type TaskOutput struct {
	TaskID int       `db:"task_id" json:"task_id"`
	Task   string    `db:"task" json:"task"`
	Time   time.Time `db:"time" json:"time"`
	Output string    `db:"output" json:"output"`
}
