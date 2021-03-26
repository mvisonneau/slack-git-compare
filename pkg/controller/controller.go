package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mvisonneau/slack-git-compare/pkg/config"
	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	"github.com/mvisonneau/slack-git-compare/pkg/providers/github"
	"github.com/mvisonneau/slack-git-compare/pkg/providers/gitlab"
	"github.com/mvisonneau/slack-git-compare/pkg/slack"
	"github.com/mvisonneau/slack-git-compare/pkg/store"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v3"
)

// Controller holds the necessary clients to run the app and handle requests
type Controller struct {
	Context        context.Context
	Providers      providers.Providers
	Store          *store.Store
	Slack          slack.Slack
	TaskController TaskController
}

// New creates a new controller
func New(ctx context.Context, cfg config.Config) (c Controller, err error) {
	c.Context = ctx
	c.Slack = slack.New(cfg.Slack, cfg.Users)
	c.Store = &store.Store{}
	c.TaskController = NewTaskController()

	err = c.configureProviders(cfg.Providers)
	if err != nil {
		return
	}

	_, _ = c.TaskController.TaskMap.Register(&taskq.TaskOptions{
		Name:    string(TaskTypeRepositoriesUpdate),
		Handler: c.TaskHandlerRepositoriesUpdate,
	})

	_, _ = c.TaskController.TaskMap.Register(&taskq.TaskOptions{
		Name:    string(TaskTypeRepositoryRefsUpdate),
		Handler: c.TaskHandlerRepositoryRefsUpdate,
	})

	_, _ = c.TaskController.TaskMap.Register(&taskq.TaskOptions{
		Name:    string(TaskTypeSlackUsersEmailsUpdate),
		Handler: c.TaskHandlerSlackUsersEmailsUpdate,
	})

	// cache updates
	c.scheduleCacheUpdateTasks(cfg.Cache)

	return
}

func (c *Controller) configureProviders(cfg config.Providers) error {
	c.Providers = make(providers.Providers)

	if len(cfg) == 0 {
		return fmt.Errorf("you must configure at least one git provider, none given")
	}

	for _, p := range cfg {
		if len(p.Owners) == 0 {
			return fmt.Errorf("you must define at least one 'owners', none given")
		}

		pt, err := providers.GetProviderTypeFromString(p.Type)
		if err != nil {
			return err
		}

		switch pt {
		case providers.ProviderTypeGitHub:
			c.Providers[pt], err = github.NewProvider(c.Context, p.Token, p.URL, p.Owners)
		case providers.ProviderTypeGitLab:
			c.Providers[pt], err = gitlab.NewProvider(p.Token, p.URL, p.Owners)
		}

		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"provider": pt.String(),
			"orgs":     p.Owners,
		}).Debug("configured provider")
	}

	return nil
}

// ScheduleTask ..
func (c Controller) ScheduleTask(tt TaskType, args ...interface{}) {
	task := c.TaskController.TaskMap.Get(string(tt))
	msg := task.WithArgs(c.Context, args...)
	if err := c.TaskController.Queue.Add(msg); err != nil {
		log.WithError(err).Warning("scheduling task")
	}
}

func (c Controller) scheduleCacheUpdateTasks(cfg config.Cache) {
	// Initialize local cache
	go func() {
		if cfg.Slack.UpdateUsersEmails.OnStart {
			c.ScheduleTask(TaskTypeSlackUsersEmailsUpdate)
		}

		wg := sync.WaitGroup{}
		if cfg.Providers.UpdateRepositories.OnStart {
			wg.Add(1)
			c.ScheduleTask(TaskTypeRepositoriesUpdate, &wg)
		}

		if cfg.Providers.UpdateRepositoriesRefs.OnStart {
			wg.Wait()
			c.ScheduleTask(TaskTypeRepositoriesRefsUpdate)
		}
	}()

	c.scheduleCacheUpdateTaskWithTicker(TaskTypeSlackUsersEmailsUpdate, cfg.Slack.UpdateUsersEmails.EverySeconds)
	c.scheduleCacheUpdateTaskWithTicker(TaskTypeRepositoriesUpdate, cfg.Providers.UpdateRepositories.EverySeconds)
	c.scheduleCacheUpdateTaskWithTicker(TaskTypeRepositoriesRefsUpdate, cfg.Providers.UpdateRepositoriesRefs.EverySeconds)
}

func (c Controller) scheduleCacheUpdateTaskWithTicker(tt TaskType, interval int) {
	if interval <= 0 {
		log.WithField("task", tt).Info("task scheduling disabled")
		return
	}

	log.WithFields(log.Fields{
		"task":             tt,
		"interval_seconds": interval,
	}).Info("task scheduled")

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for {
			select {
			case <-ctx.Done():
				log.WithField("task", tt).Info("scheduling of task stopped")
				return
			case <-ticker.C:
				switch tt {
				case TaskTypeRepositoriesUpdate:
					wg := sync.WaitGroup{}
					wg.Add(1)
					c.ScheduleTask(tt, &wg)
				default:
					c.ScheduleTask(tt)
				}
			}
		}
	}(c.Context)
}
