package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	"github.com/mvisonneau/slack-git-compare/pkg/slack"

	log "github.com/sirupsen/logrus"
	goSlack "github.com/slack-go/slack"
)

// SlashHandler handles slash command payloads
func (c Controller) SlashHandler(w http.ResponseWriter, r *http.Request) {
	err := c.Slack.VerifySigningSecret(r)
	if err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cmd, err := goSlack.SlashCommandParse(r)
	if err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch cmd.Command {
	case "/compare":
		var (
			repo           providers.Repository
			fromRef, toRef providers.Ref
			cmp            *providers.Comparison
		)

		params := strings.Split(cmd.Text, " ")
		if length := len(params); length > 0 {
			repo = c.Store.GetRepositories().GetByClosestNameMatch(params[0])
			if !repo.IsEmpty() {
				// Check if it could be worth to trigger an update of the repository's refs
				if repo.RefsLastUpdate.Add(30*time.Minute).Unix() < time.Now().Unix() || len(repo.Refs) == 0 {
					c.ScheduleTask(TaskTypeRepositoryRefsUpdate, repo.Key())
				}

				if len(repo.Refs) > 0 && length > 1 {
					fromRef = repo.Refs.GetByClosestNameMatch(params[1])
					if length > 2 {
						toRef = repo.Refs.GetByClosestNameMatch(params[2])
						if !fromRef.IsEmpty() && !toRef.IsEmpty() {
							cmp, err = c.Providers[repo.ProviderType].Compare(repo.Name, fromRef, toRef)
							if err != nil {
								log.WithError(err).Error()
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
							cmp.HydrateCommitsAuthorsWithSlackUserID(c.Store.GetSlackUsersEmails())
						}
					}
				}
			}
		}

		modalRequest := slack.GenerateModalRequestRepositoryPicker(cmd.ChannelID, repo, c.Store.GetRepositoriesLastUpdate(), fromRef, toRef, repo.RefsLastUpdate, cmp)
		resp, err := c.Slack.Client.OpenView(cmd.TriggerID, modalRequest)
		if err != nil {
			log.WithError(fmt.Errorf("opening view: %s -> %v", err.Error(), resp.ResponseMetadata)).Error()
		}
	default:
		log.WithError(fmt.Errorf("unhandled command '%s'", cmd.Command)).Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// ModalHandler handlers slack modal payloads
func (c Controller) ModalHandler(w http.ResponseWriter, r *http.Request) {
	err := c.Slack.VerifySigningSecret(r)
	if err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var i goSlack.InteractionCallback
	err = json.Unmarshal([]byte(r.FormValue("payload")), &i)
	if err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// If no state values are being passed, it means it has probably be a link being clicked
	// We simply ignore the call.
	if i.View.State != nil {
		if _, ok := i.View.State.Values["repositories"]["repository"]; !ok {
			log.Debug("ignoring call as no state values are being defined")
			return
		}
	} else {
		log.Debug("ignoring call as no state values are being defined")
		return
	}

	var (
		repo           providers.Repository
		fromRef, toRef providers.Ref
		cmp            *providers.Comparison
	)

	repoKey := i.View.State.Values["repositories"]["repository"].SelectedOption.Value
	if len(repoKey) > 0 {
		var found bool
		repo, found = c.Store.GetRepository(providers.RepositoryKey(stripRankFromValue(repoKey)))
		if !found {
			log.WithField("repository_key", stripRankFromValue(repoKey)).WithError(fmt.Errorf("repository not found")).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	fromRefKey := i.View.State.Values["from_ref"]["from_ref"].SelectedOption.Value
	if len(fromRefKey) > 0 {
		var found bool
		fromRef, found = repo.Refs[providers.RefKey(stripRankFromValue(fromRefKey))]
		if !found {
			if err = c.Slack.SendMessage(i.View.CallbackID, fmt.Sprintf("unable to find ref_key `%s`", stripRankFromValue(fromRefKey))); err != nil {
				log.WithError(err).Error()
			}
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	toRefKey := i.View.State.Values["to_ref"]["to_ref"].SelectedOption.Value
	if len(toRefKey) > 0 {
		var found bool
		toRef, found = repo.Refs[providers.RefKey(stripRankFromValue(toRefKey))]
		if !found {
			if err = c.Slack.SendMessage(i.View.CallbackID, fmt.Sprintf("unable to find ref_key `%s`", stripRankFromValue(toRefKey))); err != nil {
				log.WithError(err).Error()
			}
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if !repo.IsEmpty() && !fromRef.IsEmpty() && !toRef.IsEmpty() {
		cmp, err = c.Providers[repo.ProviderType].Compare(repo.Name, fromRef, toRef)
		if err != nil {
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cmp.HydrateCommitsAuthorsWithSlackUserID(c.Store.GetSlackUsersEmails())
	}

	// We only want to update the view when we change the repository select
	switch i.Type {
	case goSlack.InteractionTypeBlockActions:
		resp, err := c.Slack.Client.UpdateView(slack.GenerateModalRequestRepositoryPicker(i.View.CallbackID, repo, c.Store.GetRepositoriesLastUpdate(), fromRef, toRef, repo.RefsLastUpdate, cmp), "", i.View.Hash, i.View.ID)
		if err != nil {
			log.WithError(fmt.Errorf("updating view: %s -> %v", err.Error(), resp.ResponseMetadata)).Error()
		}
	case goSlack.InteractionTypeViewSubmission:
		if repo.IsEmpty() {
			log.Error("did not expect repo to be undefined")
			return
		}
		if _, _, err := c.Slack.Client.PostMessage(i.View.CallbackID, goSlack.MsgOptionBlocks(slack.GenerateComparisonMessage(repo, fromRef, toRef, *cmp, i.User.ID).BlockSet...)); err != nil {
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	for _, a := range i.ActionCallback.BlockActions {
		if a != nil {
			switch a.ActionID {
			case "update_repositories":
				log.Info("triggered an update of the repositories list")
				c.ScheduleTask(TaskTypeRepositoriesUpdate)
			case "update_refs":
				log.WithField("repository_key", a.Value).Info("triggered an update of the refs list")
				c.ScheduleTask(TaskTypeRepositoryRefsUpdate, a.Value)
			}
		}
	}
}

// SelectHandler handles slack selector payloads
func (c Controller) SelectHandler(w http.ResponseWriter, r *http.Request) {
	i := &goSlack.InteractionCallback{}
	if err := json.Unmarshal([]byte(r.FormValue("payload")), i); err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"filter": i.Value,
		"action": i.ActionID,
	}).Debug("selector search")

	resp := goSlack.OptionsResponse{}
	switch i.ActionID {
	case "repository":
		for _, r := range c.Store.GetRepositories().Search(i.Value, 20) {
			resp.Options = append(resp.Options, goSlack.NewOptionBlockObject(fmt.Sprintf("%d/%s", r.Rank, r.Key()), goSlack.NewTextBlockObject("plain_text", fmt.Sprintf(":%s: %s", r.ProviderType, r.Name), true, false), nil))
		}
	case "from_ref", "to_ref":
		repoKey := providers.RepositoryKey(i.View.PrivateMetadata)
		repo, found := c.Store.GetRepository(repoKey)
		if !found {
			log.WithField("repository_key", repoKey).WithError(fmt.Errorf("repository not found")).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check if it could be worth to trigger an update of the repository's refs
		if repo.RefsLastUpdate.Add(30*time.Minute).Unix() < time.Now().Unix() || len(repo.Refs) == 0 {
			c.ScheduleTask(TaskTypeRepositoryRefsUpdate, repo.Key())
		}

		for _, r := range repo.Refs.Search(i.Value, 20) {
			resp.Options = append(resp.Options, goSlack.NewOptionBlockObject(fmt.Sprintf("%d/%s", r.Rank, r.Key()), goSlack.NewTextBlockObject("plain_text", fmt.Sprintf("%s/%s", r.Type, r.Name), true, false), nil))
		}
	default:
		log.WithField("action_id", i.ActionID).Error("unsupported action_id")
	}

	jsonResp, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.WithError(err).Error()
	}
}

func stripRankFromValue(value string) string {
	values := strings.Split(value, "/")
	if len(values) != 2 {
		return ""
	}
	return values[1]
}
