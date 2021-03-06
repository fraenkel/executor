package runoncehandler

import (
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	steno "github.com/cloudfoundry/gosteno"
	"github.com/vito/gordon"

	"github.com/cloudfoundry-incubator/executor/action_runner"
	"github.com/cloudfoundry-incubator/executor/actionrunner"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/claim_action"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/complete_action"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/create_container_action"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/execute_action"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/register_action"
	"github.com/cloudfoundry-incubator/executor/taskregistry"
)

type RunOnceHandlerInterface interface {
	RunOnce(runOnce models.RunOnce, executorId string)
}

type RunOnceHandler struct {
	bbs          Bbs.ExecutorBBS
	wardenClient gordon.Client
	actionRunner actionrunner.ActionRunnerInterface

	loggregatorServer string
	loggregatorSecret string

	logger *steno.Logger

	taskRegistry taskregistry.TaskRegistryInterface

	stack string
}

func New(
	bbs Bbs.ExecutorBBS,
	wardenClient gordon.Client,
	taskRegistry taskregistry.TaskRegistryInterface,
	actionRunner actionrunner.ActionRunnerInterface,
	loggregatorServer string,
	loggregatorSecret string,
	stack string,
	logger *steno.Logger,
) *RunOnceHandler {
	return &RunOnceHandler{
		bbs:               bbs,
		wardenClient:      wardenClient,
		taskRegistry:      taskRegistry,
		actionRunner:      actionRunner,
		loggregatorServer: loggregatorServer,
		loggregatorSecret: loggregatorSecret,
		logger:            logger,
		stack:             stack,
	}
}

func (handler *RunOnceHandler) RunOnce(runOnce models.RunOnce, executorID string) {
	// check for stack compatibility
	// move to task registry?
	if runOnce.Stack != "" && handler.stack != runOnce.Stack {
		handler.logger.Errord(map[string]interface{}{"runonce-guid": runOnce.Guid, "desired-stack": runOnce.Stack, "executor-stack": handler.stack}, "runonce.stack.mismatch")
		return
	}

	runner := action_runner.New([]action_runner.Action{
		register_action.New(
			runOnce,
			handler.logger,
			handler.taskRegistry,
		),
		claim_action.New(
			&runOnce,
			handler.logger,
			executorID,
			handler.bbs,
		),
		create_container_action.New(
			&runOnce,
			handler.logger,
			handler.wardenClient,
		),
		execute_action.New(
			&runOnce,
			handler.logger,
			handler.bbs,
			handler.actionRunner,
			handler.loggregatorServer,
			handler.loggregatorSecret,
		),
		complete_action.New(
			&runOnce,
			handler.logger,
			handler.bbs,
		),
	})

	result := make(chan error, 1)

	go runner.Perform(result)

	<-result
}
