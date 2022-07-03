package task

import (
	"fmt"
	"net/http"

	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	task_model "code.gitea.io/gitea/models/task"
)

const (
	tplMiniTask    base.TplName = "task/minitask"
)

// MiniTask is the minitask page
func MiniTask(ctx *context.Context) {
	var tasks []*task_model.MiniTask

	var err error
	if tasks, err = task_model.GetAllTasks(); err != nil {
		ctx.ServerError("xxxxxxxxxxxxxxxxx: ", err)
		return
	}

	for _, task := range tasks {
		fmt.Printf("xxxxxxxxxxxxxxxxxxxxx%s", task.ID)
	}


	ctx.Data["PageIsMiniTask"] = true
	ctx.HTML(http.StatusOK, tplMiniTask)
}

