package db

type Template struct {
	ID int `db:"id" json:"id"`

	SshKeyID      int  `db:"ssh_key_id" json:"ssh_key_id"`
	ProjectID     int  `db:"project_id" json:"project_id"`
	InventoryID   int  `db:"inventory_id" json:"inventory_id"`
	RepositoryID  int  `db:"repository_id" json:"repository_id"`
	EnvironmentID *int `db:"environment_id" json:"environment_id"`
	BuildTemplateID *int `db:"build_template_id" json:"build_template_id"`

	// Alias as described in https://github.com/ansible-semaphore/semaphore/issues/188
	Alias string `db:"alias" json:"alias"`
	Type *string `db:"type" json:"type"`

	// playbook name in the form of "some_play.yml"
	Playbook string `db:"playbook" json:"playbook"`
	// to fit into []string
	Arguments *string `db:"arguments" json:"arguments"`
	// if true, semaphore will not prepend any arguments to `arguments` like inventory, etc
	OverrideArguments bool `db:"override_args" json:"override_args"`
	Removed   bool    `db:"removed" json:"removed"`
	LastSuccessTaskID *int `db:"last_success_task_id" json:"last_success_task_id"`
	LastSuccessBuildTaskID *int `db:"last_success_build_task_id" json:"last_success_build_task_id"`
	VersionTemplate *string `db:"version_template" json:"version_template"`
}

type TemplateSchedule struct {
	TemplateID int    `db:"template_id" json:"template_id"`
	CronFormat string `db:"cron_format" json:"cron_format"`
}
