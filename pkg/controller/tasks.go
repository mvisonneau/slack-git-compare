package controller

import (
	"fmt"
	"sync"
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

	// TaskTypeRepositoriesRefsUpdate updates all Repositories in the local store with refs fetched from
	// their associated git provider
	TaskTypeRepositoriesRefsUpdate TaskType = "RepositoriesRefsUpdate"

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
func (c *Controller) TaskHandlerRepositoriesUpdate(wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	if c.Store.GetRepositoriesLastUpdate().Add(time.Minute).Unix() > time.Now().Unix() {
		log.Debug("repositories updated less than a minute ago, skipping..")
		return
	}

	repos, err := c.Providers.ListRepositories()
	if err != nil {
		log.WithError(err).Warning("executing 'RepositoriesUpdate' task")
		return
	}

	c.Store.UpdateRepositories(repos)
	log.Info("updated repositories list")
	return
}

// TaskHandlerRepositoriesRefsUpdate updates all Repositories in the local store with refs fetched from
// its associated git provider
func (c *Controller) TaskHandlerRepositoriesRefsUpdate() {
	for _, r := range c.Store.GetRepositories() {
		var err error
		r.Refs, err = c.Providers[r.ProviderType].ListRefs(r.Name)
		if err != nil {
			log.WithError(err).Warning("executing 'RepositoriesRefsUpdate' task")
			return
		}

		r.RefsLastUpdate = time.Now()
		c.Store.UpdateRepository(r)
		log.WithFields(log.Fields{
			"repository_provider": r.ProviderType,
			"repository_name":     r.Name,
		}).Info("updated repo refs list!")
	}

	return
}

// TaskHandlerRepositoryRefsUpdate updates a Repository in the local store with refs fetched from
// its associated git provider
func (c *Controller) TaskHandlerRepositoryRefsUpdate(wg *sync.WaitGroup, rk providers.RepositoryKey) {
	if wg != nil {
		defer wg.Done()
	}

	r, found := c.Store.GetRepository(rk)
	if !found {
		err := fmt.Errorf("repository key '%s' not found in store", rk)
		log.WithError(err).WithField("repository_key", rk).Warning("executing 'RepositoryRefsUpdate' task")
		return
	}

	if r.RefsLastUpdate.Add(time.Minute).Unix() > time.Now().Unix() {
		log.Debug("refs updated less than a minute ago, skipping..")
		return
	}

	var err error
	r.Refs, err = c.Providers[r.ProviderType].ListRefs(r.Name)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"repository_provider": r.ProviderType,
			"repository_name":     r.Name,
		}).Warning("executing 'RepositoryRefsUpdate' task")
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
func (c *Controller) TaskHandlerSlackUsersEmailsUpdate() {
	if c.Store.GetSlackUsersEmailsLastUpdate().Add(time.Minute).Unix() > time.Now().Unix() {
		log.Debug("slack users emails updated less than a minute ago, skipping..")
		return
	}

	sue, err := c.Slack.ListSlackUserEmailMappings()
	if err != nil {
		log.WithError(err).Warning("executing 'SlackUsersEmailsUpdate' task")
		return
	}

	c.Store.UpdateSlackUsersEmails(sue)
	log.Info("updated slack users emails mapping list")
	return
}
