package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

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
		opts := slack.ModalRequestOptions{
			ConversationID:         cmd.ChannelID,
			LastRepositoriesUpdate: c.Store.GetRepositoriesLastUpdate(),
		}

		if len(c.Store.GetRepositories()) == 0 ||
			opts.LastRepositoriesUpdate.IsZero() {
			opts.CurrentlyUpdatingRepositories = true
		} else {
			params := strings.Split(cmd.Text, " ")
			if length := len(params); length > 0 {
				opts.Repository = c.Store.GetRepositories().GetByClosestNameMatch(params[0])
				if !opts.Repository.IsEmpty() {
					// Check if it could be worth to trigger an update of the repository's refs
					if opts.Repository.RefsLastUpdate.IsZero() ||
						len(opts.Repository.Refs) == 0 {
						opts.CurrentlyUpdatingRepositoryRefs = true
					}

					if len(opts.Repository.Refs) > 0 && length > 1 {
						opts.FromRef = opts.Repository.Refs.GetByClosestNameMatch(params[1])
						if length > 2 {
							opts.ToRef = opts.Repository.Refs.GetByClosestNameMatch(params[2])
							if !opts.FromRef.IsEmpty() && !opts.ToRef.IsEmpty() {
								opts.Comparison, err = c.Providers[opts.Repository.ProviderType].Compare(opts.Repository.Name, opts.FromRef, opts.ToRef)
								if err != nil {
									log.WithError(err).Error()
									w.WriteHeader(http.StatusInternalServerError)
									return
								}
								opts.Comparison.HydrateCommitsAuthorsWithSlackUserID(c.Store.GetSlackUsersEmails())
							}
						}
					}
				}
			}
		}

		// Generate and submit the request to open the modal
		resp, err := c.Slack.Client.OpenView(
			cmd.TriggerID,
			slack.GetModalRequest(opts),
		)
		if err != nil {
			log.WithError(fmt.Errorf("opening view: %s -> %v", err.Error(), resp.ResponseMetadata)).Error()
			return
		}

		c.handleRequiredDataFetchesAndUpdateModalAfterCompletion(resp.ID, resp.Hash, opts)
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

	opts := slack.ModalRequestOptions{
		ConversationID:         i.View.CallbackID,
		LastRepositoriesUpdate: c.Store.GetRepositoriesLastUpdate(),
	}

	if len(c.Store.GetRepositories()) == 0 ||
		opts.LastRepositoriesUpdate.IsZero() {
		opts.CurrentlyUpdatingRepositories = true
	} else {
		repoKey := i.View.State.Values["repositories"]["repository"].SelectedOption.Value
		if len(repoKey) > 0 {
			var found bool
			opts.Repository, found = c.Store.GetRepository(providers.RepositoryKey(stripRankFromValue(repoKey)))
			if !found {
				log.WithField("repository_key", stripRankFromValue(repoKey)).WithError(fmt.Errorf("repository not found")).Error()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Check if it could be worth to trigger an update of the repository's refs
			if opts.Repository.RefsLastUpdate.IsZero() ||
				len(opts.Repository.Refs) == 0 {
				opts.CurrentlyUpdatingRepositoryRefs = true
			}
		}

		fromRefKey := i.View.State.Values["from_ref"]["from_ref/"+string(opts.Repository.Key())].SelectedOption.Value
		if len(fromRefKey) > 0 {
			opts.FromRef, _ = opts.Repository.Refs[providers.RefKey(stripRankFromValue(fromRefKey))]
			// TODO: If last updated is quite old and ref is not found, trigger an update of the refs
		}

		toRefKey := i.View.State.Values["to_ref"]["to_ref/"+string(opts.Repository.Key())].SelectedOption.Value
		if len(toRefKey) > 0 {
			opts.ToRef, _ = opts.Repository.Refs[providers.RefKey(stripRankFromValue(toRefKey))]
			// TODO: If last updated is quite old and ref is not found, trigger an update of the refs
		}

		if !opts.Repository.IsEmpty() &&
			!opts.FromRef.IsEmpty() &&
			!opts.ToRef.IsEmpty() {
			log.Debug("comparing refs")
			opts.Comparison, err = c.Providers[opts.Repository.ProviderType].Compare(opts.Repository.Name, opts.FromRef, opts.ToRef)
			if err != nil {
				log.WithError(err).Error()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			opts.Comparison.HydrateCommitsAuthorsWithSlackUserID(c.Store.GetSlackUsersEmails())
		}
	}

	for _, a := range i.ActionCallback.BlockActions {
		if a != nil {
			switch a.ActionID {
			case "update_repositories":
				log.Info("triggered an update of the repositories list")
				opts.CurrentlyUpdatingRepositories = true
			case "update_refs":
				log.WithField("repository_key", a.Value).Info("triggered an update of the refs list")
				var found bool
				opts.Repository, found = c.Store.GetRepository(providers.RepositoryKey(a.Value))
				if !found {
					log.WithField("repository_key", a.Value).WithError(fmt.Errorf("repository not found")).Error()
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				opts.CurrentlyUpdatingRepositoryRefs = true
			}
		}
	}

	// We only want to update the view when we change the repository select
	switch i.Type {
	case goSlack.InteractionTypeBlockActions:
		resp, err := c.Slack.Client.UpdateView(slack.GetModalRequest(opts), "", i.View.Hash, i.View.ID)
		if err != nil {
			log.WithError(fmt.Errorf("updating view: %s -> %v", err.Error(), resp.ResponseMetadata)).Error()
		}
		c.handleRequiredDataFetchesAndUpdateModalAfterCompletion(resp.ID, resp.Hash, opts)
	case goSlack.InteractionTypeViewSubmission:
		errors := map[string]string{}

		// It does not seem to actually reflect in the modal errors but at least
		// it seems to work..
		if opts.Repository.IsEmpty() {
			errors["repositories"] = "Please select a repository"
		} else {
			if opts.FromRef.IsEmpty() {
				errors["from_ref"] = "Please select a base ref"
			}
			if opts.FromRef.IsEmpty() {
				errors["to_ref"] = "Please select a head ref"
			}
		}

		if len(errors) > 0 {
			vsr := slack.ViewSubmissionResponse{
				ResponseType: "errors",
				Errors:       errors,
			}

			resp, _ := json.Marshal(vsr)
			w.Write(resp)
			return
		}

		if opts.Comparison == nil {
			log.Error("comparison was not available between the 2 provided refs")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, _, err := c.Slack.Client.PostMessage(i.View.CallbackID, goSlack.MsgOptionBlocks(slack.GenerateComparisonMessage(opts.Repository, opts.FromRef, opts.ToRef, *opts.Comparison, i.User.ID).BlockSet...)); err != nil {
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		log.Warning("unsupported interaction type '%s'", i.Type)
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

	actionID := i.ActionID
	var repoKey providers.RepositoryKey
	if strings.Contains(actionID, "/") {
		values := strings.Split(actionID, "/")
		actionID = values[0]
		repoKey = providers.RepositoryKey(values[1])
	}

	log.WithFields(log.Fields{
		"filter": i.Value,
		"action": actionID,
	}).Debug("selector search")

	resp := goSlack.OptionsResponse{}
	switch actionID {
	case "repository":
		for _, r := range c.Store.GetRepositories().Search(i.Value, 20) {
			resp.Options = append(resp.Options, goSlack.NewOptionBlockObject(fmt.Sprintf("%d/%s", r.Rank, r.Key()), goSlack.NewTextBlockObject("plain_text", fmt.Sprintf(":%s: %s", r.ProviderType, r.Name), true, false), nil))
		}
	case "from_ref", "to_ref":
		repo, found := c.Store.GetRepository(repoKey)
		if !found {
			log.WithField("repository_key", repoKey).WithError(fmt.Errorf("repository not found")).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
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

func (c Controller) handleRequiredDataFetchesAndUpdateModalAfterCompletion(viewID, viewHash string, opts slack.ModalRequestOptions) {
	if opts.CurrentlyUpdatingRepositories {
		go func() {
			wg := sync.WaitGroup{}
			wg.Add(1)
			c.ScheduleTask(TaskTypeRepositoriesUpdate, &wg)
			wg.Wait()

			opts.CurrentlyUpdatingRepositories = false
			opts.LastRepositoriesUpdate = c.Store.GetRepositoriesLastUpdate()
			r, err := c.Slack.Client.UpdateView(slack.GetModalRequest(opts), "", viewHash, viewID)
			if err != nil {
				log.WithError(fmt.Errorf("updating view: %s -> %v", err.Error(), r.ResponseMetadata)).Error()
			}
		}()
	}

	if opts.CurrentlyUpdatingRepositoryRefs {
		go func() {
			wg := sync.WaitGroup{}
			wg.Add(1)
			c.ScheduleTask(TaskTypeRepositoryRefsUpdate, opts.Repository.Key(), &wg)
			wg.Wait()

			opts.CurrentlyUpdatingRepositoryRefs = false
			opts.Repository, _ = c.Store.GetRepository(opts.Repository.Key())
			r, err := c.Slack.Client.UpdateView(slack.GetModalRequest(opts), "", viewHash, viewID)
			if err != nil {
				log.WithError(fmt.Errorf("updating view: %s -> %v", err.Error(), r.ResponseMetadata)).Error()
			}
		}()
	}
}
