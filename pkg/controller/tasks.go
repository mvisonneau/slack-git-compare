package controller

import (
	"fmt"
	"time"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
)

// TaskController holds task related clients
type TaskController struct {
	Factory taskq.Factory
	Queue   taskq.Queue
	TaskMap *taskq.TaskMap
}

// TaskType represents the type of a task
type TaskType string

const (
	// TaskTypeRepositoriesUpdate updates the local store with repositories fetched from
	// configured git providers
	TaskTypeRepositoriesUpdate TaskType = "RepositoriesUpdate"

	// TaskTypeRepositoryRefsUpdate updates a Repository in the local store with refs fetched from
	// its associated git provider
	TaskTypeRepositoryRefsUpdate TaskType = "RepositoryRefsUpdate"

	// TaskTypeSlackUsersEmailsUpdate updates the local store with slack users emails fetched from
	// the Slack API and local configuration (for custom aliases)
	TaskTypeSlackUsersEmailsUpdate TaskType = "SlackUsersEmailsUpdate"
)

// NewTaskController initializes and returns a new TaskController object
func NewTaskController() (t TaskController) {
	t.TaskMap = &taskq.TaskMap{}
	t.Factory = memqueue.NewFactory()
	t.Queue = t.Factory.RegisterQueue(&taskq.QueueOptions{
		Name:                 "default",
		PauseErrorsThreshold: 3,
		Handler:              t.TaskMap,

		// Disable system resources checks
		MinSystemResources: taskq.SystemResources{
			Load1PerCPU:          -1,
			MemoryFreeMB:         0,
			MemoryFreePercentage: 0,
		},
	})

	return
}

// TaskHandlerRepositoriesUpdate updates the local store with repositories fetched from
// configured git providers
func (c *Controller) TaskHandlerRepositoriesUpdate() (err error) {
	if c.Store.GetRepositoriesLastUpdate().Add(2*time.Minute).Unix() > time.Now().Unix() {
		log.Debug("repositories updated less than 2 minutes ago, skipping..")
		return
	}

	repos, err := c.Providers.ListRepositories()
	if err != nil {
		return err
	}

	c.Store.UpdateRepositories(repos)
	log.Info("updated repositories list")
	return
}

// TaskHandlerRepositoryRefsUpdate updates a Repository in the local store with refs fetched from
// its associated git provider
func (c *Controller) TaskHandlerRepositoryRefsUpdate(rk providers.RepositoryKey) (err error) {
	r, found := c.Store.GetRepository(rk)
	if !found {
		return fmt.Errorf("repository key '%s' not found in store", rk)
	}

	if r.RefsLastUpdate.Add(2*time.Minute).Unix() > time.Now().Unix() {
		log.Debug("refs updated less than 2 minutes ago, skipping..")
		return
	}

	r.Refs, err = c.Providers[r.ProviderType].ListRefs(r.Name)
	if err != nil {
		return
	}
	r.RefsLastUpdate = time.Now()

	c.Store.UpdateRepository(r)
	log.WithFields(log.Fields{
		"repository_provider": r.ProviderType,
		"repository_name":     r.Name,
	}).Info("updated repo refs list!")
	return
}

// TaskHandlerSlackUsersEmailsUpdate updates the local store with slack users emails fetched from
// the Slack API and local configuration (for custom aliases)
func (c *Controller) TaskHandlerSlackUsersEmailsUpdate() (err error) {
	if c.Store.GetSlackUsersEmailsLastUpdate().Add(2*time.Minute).Unix() > time.Now().Unix() {
		log.Debug("slack users emails updated less than 2 minutes ago, skipping..")
		return
	}

	sue, err := c.Slack.ListSlackUserEmailMappings()
	if err != nil {
		return err
	}

	c.Store.UpdateSlackUsersEmails(sue)
	log.Info("updated slack users emails mapping list")
	return
}
