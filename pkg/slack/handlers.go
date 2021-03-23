package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	"github.com/openlyinc/pointy"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// This was taken from the slash example
// https://github.com/slack-go/slack/blob/master/examples/slash/slash.go
func (s Slack) verifySigningSecret(r *http.Request) error {
	verifier, err := slack.NewSecretsVerifier(r.Header, s.SigningSecret)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Need to use r.Body again when unmarshalling SlashCommand and InteractionCallback
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if _, err = verifier.Write(body); err != nil {
		return err
	}

	return verifier.Ensure()
}

func generateModalRequestRepositoryPicker(conversationID string, repo *providers.Repository, fromRef, toRef *providers.Ref, cmp *providers.Comparison) slack.ModalViewRequest {
	modalRequest := slack.ModalViewRequest{}
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = slack.NewTextBlockObject(slack.PlainTextType, "git compare", false, false)
	modalRequest.Close = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	modalRequest.Submit = slack.NewTextBlockObject(slack.PlainTextType, "Post to channel", false, false)
	modalRequest.CallbackID = conversationID

	// repositoriesRefreshSection := slack.NewSectionBlock(
	// 	slack.NewTextBlockObject(slack.MarkdownType, "_last updated 2m ago_", false, false),
	// 	nil,
	// 	slack.NewAccessory(
	// 		slack.NewButtonBlockElement(
	// 			"refresh_repositories",
	// 			"",
	// 			slack.NewTextBlockObject(
	// 				slack.PlainTextType,
	// 				"update now",
	// 				false,
	// 				false,
	// 			),
	// 		),
	// 	),
	// )

	repositoriesElement := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, "repository")
	repositoriesElement.MinQueryLength = pointy.Int(0)
	repositoriesInput := slack.NewInputBlock(
		"repositories",
		slack.NewTextBlockObject(slack.PlainTextType, "Select a repository", false, false),
		repositoriesElement,
	)
	repositoriesInput.DispatchAction = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			repositoriesInput,
			// repositoriesRefreshSection,
		},
	}

	if repo != nil {
		// This is how we keep the selected repository across modal refreshes
		repositoriesElement.InitialOption = slack.NewOptionBlockObject(fmt.Sprintf("x/%s", repo.Key()), slack.NewTextBlockObject("plain_text", fmt.Sprintf(":%s: %s", repo.ProviderType, repo.Name), true, false), nil)

		// When need to store the repository key in here to be able to read it
		// during the block submission payloads of to_ref & from_ref external_select
		modalRequest.PrivateMetadata = string(repo.Key())

		// Add a divider
		blocks.BlockSet = append(blocks.BlockSet, slack.NewDividerBlock())

		// FROM REF
		fromRefElement := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, "from_ref")
		fromRefElement.MinQueryLength = pointy.Int(0)
		if fromRef != nil {
			fromRefElement.InitialOption = slack.NewOptionBlockObject(
				fmt.Sprintf("x/%s", fromRef.Key()),
				slack.NewTextBlockObject(
					slack.PlainTextType,
					fmt.Sprintf("%s/%s", fromRef.Type, fromRef.Name),
					true,
					false,
				),
				nil,
			)
		}

		fromRefInput := slack.NewInputBlock(
			"from_ref",
			slack.NewTextBlockObject(slack.PlainTextType, "From (BASE)", false, false),
			fromRefElement,
		)
		fromRefInput.DispatchAction = true

		// TO REF
		toRefElement := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, "to_ref")
		toRefElement.MinQueryLength = pointy.Int(0)
		if toRef != nil {
			toRefElement.InitialOption = slack.NewOptionBlockObject(
				fmt.Sprintf("x/%s", toRef.Key()),
				slack.NewTextBlockObject(
					slack.PlainTextType,
					fmt.Sprintf("%s/%s", toRef.Type, toRef.Name),
					true,
					false,
				),
				nil,
			)
		}

		toRefInput := slack.NewInputBlock(
			"to_ref",
			slack.NewTextBlockObject(slack.PlainTextType, "To (HEAD)", false, false),
			toRefElement,
		)
		toRefInput.DispatchAction = true

		// Refresh refs
		// refsRefreshSection := slack.NewSectionBlock(
		// 	slack.NewTextBlockObject(slack.MarkdownType, "_last updated 2m ago_", false, false),
		// 	nil,
		// 	slack.NewAccessory(
		// 		slack.NewButtonBlockElement(
		// 			"refresh_refs",
		// 			string(repo.Key()),
		// 			slack.NewTextBlockObject(
		// 				slack.PlainTextType,
		// 				"update now",
		// 				false,
		// 				false,
		// 			),
		// 		),
		// 	),
		// )

		// Add the refs selectors
		blocks.BlockSet = append(blocks.BlockSet, fromRefInput)
		blocks.BlockSet = append(blocks.BlockSet, toRefInput)
		// blocks.BlockSet = append(blocks.BlockSet, refsRefreshSection)

		if cmp != nil {
			// Add a divider
			blocks.BlockSet = append(blocks.BlockSet, slack.NewDividerBlock())

			var msg string
			if cmp.CommitCount() == 0 {
				msg = ":shrug: there are no difference between the refs"
			} else {
				commitString := "commit"
				if cmp.CommitCount() > 1 {
					commitString += "s"
				}
				msg = fmt.Sprintf("*%d %s* found the between the refs\n%s", cmp.CommitCount(), commitString, strings.TrimPrefix(cmp.AuthorsSlackString(), commitString+" "))
			}

			headerButtonText := fmt.Sprintf("View in %s", repo.ProviderType.StringPretty())
			headerButton := slack.NewButtonBlockElement("", "", slack.NewTextBlockObject("plain_text", headerButtonText, false, false))
			headerButton.URL = cmp.WebURL

			blocks.BlockSet = append(blocks.BlockSet, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", msg, false, false), nil, slack.NewAccessory(headerButton)))
		}
	}

	modalRequest.Blocks = blocks

	return modalRequest
}

func generateComparisonMessage(repo providers.Repository, fromRef, toRef providers.Ref, cmp providers.Comparison, slackUserID string) slack.Blocks {
	headerText := fmt.Sprintf(
		":%s: *<%s|%s>*\n`%s/%s` :arrow_right: `%s/%s`",
		repo.ProviderType,
		repo.WebURL,
		repo.Name,
		fromRef.Type,
		fromRef.Name,
		toRef.Type,
		toRef.Name,
	)

	headerButtonText := fmt.Sprintf("View in %s", repo.ProviderType.StringPretty())
	headerButton := slack.NewButtonBlockElement("", "", slack.NewTextBlockObject("plain_text", headerButtonText, false, false))
	headerButton.URL = cmp.WebURL

	var commitsText string
	if len(cmp.Commits) == 0 {
		commitsText += ":shrug: there are no difference between the refs\n"
	}

	if len(cmp.Commits) > 10 {
		commitsText += fmt.Sprintf(":warning: there are %d commits between these refs, truncating to the 10 most recent ones\n", len(cmp.Commits))
	}

	for i, c := range cmp.Commits {
		if i > 10 {
			break
		}

		commitsText += fmt.Sprintf("> <%s|%s> | %s\n> _%s_\n", c.WebURL, c.ShortID, c.AuthorSlackString(), c.ShortMessage())
	}

	footerText := fmt.Sprintf("comparison requested by <@%s> | %s", slackUserID, cmp.AuthorsSlackString())
	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", headerText, false, false), nil, slack.NewAccessory(headerButton)),
			slack.NewDividerBlock(),
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", commitsText, false, false), nil, nil),
			slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", footerText, false, false)),
		},
	}

	return blocks
}

// SlashHandler handles slash command payloads
func (s Slack) SlashHandler(w http.ResponseWriter, r *http.Request) {
	err := s.verifySigningSecret(r)
	if err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cmd, err := slack.SlashCommandParse(r)
	if err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch cmd.Command {
	case "/compare":
		var (
			repo           *providers.Repository
			fromRef, toRef *providers.Ref
			cmp            *providers.Comparison
		)

		go func() {
			params := strings.Split(cmd.Text, " ")
			if length := len(params); length > 0 {
				repo = s.Storage.Repositories.GetByClosestNameMatch(params[0])
				if repo != nil {
					if repo.LastRefsUpdate.Add(time.Minute).Unix() < time.Now().Unix() || len(repo.Refs) == 0 {
						repo.Refs, err = s.Providers[repo.ProviderType].ListRefs(repo.Name)
						if err != nil {
							log.WithField("repository_key", repo.Key()).WithError(err).Error()
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
						repo.LastRefsUpdate = time.Now()
					}

					if length > 1 {
						fromRef = repo.Refs.GetByClosestNameMatch(params[1])
						if length > 2 {
							toRef = repo.Refs.GetByClosestNameMatch(params[2])
							if fromRef != nil && toRef != nil {
								cmp, err = s.Providers[repo.ProviderType].Compare(repo.Name, *fromRef, *toRef)
								if err != nil {
									log.WithError(err).Error()
									w.WriteHeader(http.StatusInternalServerError)
									return
								}
								cmp.HydrateCommitsAuthorsWithSlackUserID(s.Storage.SlackUserEmailMappings)
							}
						}
					}
				}
			}

			modalRequest := generateModalRequestRepositoryPicker(cmd.ChannelID, repo, fromRef, toRef, cmp)
			resp, err := s.Client.OpenView(cmd.TriggerID, modalRequest)
			if err != nil {
				log.WithError(fmt.Errorf("opening view: %s -> %v", err.Error(), resp.ResponseMetadata)).Error()
			}
		}()
	default:
		log.WithError(fmt.Errorf("unhandled command '%s'", cmd.Command)).Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// ModalHandler handlers slack modal payloads
func (s Slack) ModalHandler(w http.ResponseWriter, r *http.Request) {
	err := s.verifySigningSecret(r)
	if err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var i slack.InteractionCallback
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
		repo           *providers.Repository
		fromRef, toRef *providers.Ref
		cmp            *providers.Comparison
	)

	repoKey := i.View.State.Values["repositories"]["repository"].SelectedOption.Value
	if len(repoKey) > 0 {
		var found bool
		repo, found = s.Storage.Repositories.GetByKey(providers.RepositoryKey(stripRankFromValue(repoKey)))
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
			if err = s.sendMessage(i.View.CallbackID, fmt.Sprintf("unable to find ref_key `%s`", stripRankFromValue(fromRefKey))); err != nil {
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
			if err = s.sendMessage(i.View.CallbackID, fmt.Sprintf("unable to find ref_key `%s`", stripRankFromValue(toRefKey))); err != nil {
				log.WithError(err).Error()
			}
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if repo != nil && toRef != nil && fromRef != nil {
		cmp, err = s.Providers[repo.ProviderType].Compare(repo.Name, *fromRef, *toRef)
		if err != nil {
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cmp.HydrateCommitsAuthorsWithSlackUserID(s.Storage.SlackUserEmailMappings)
	}

	// We only want to update the view when we change the repository select
	switch i.Type {
	case slack.InteractionTypeBlockActions:
		resp, err := s.Client.UpdateView(generateModalRequestRepositoryPicker(i.View.CallbackID, repo, fromRef, toRef, cmp), "", i.View.Hash, i.View.ID)
		if err != nil {
			log.WithError(fmt.Errorf("updating view: %s -> %v", err.Error(), resp.ResponseMetadata)).Error()
		}
	case slack.InteractionTypeViewSubmission:
		if repo == nil {
			log.Error("did not expect repo to be nil")
			return
		}
		if _, _, err := s.Client.PostMessage(i.View.CallbackID, slack.MsgOptionBlocks(generateComparisonMessage(*repo, *fromRef, *toRef, *cmp, i.User.ID).BlockSet...)); err != nil {
			log.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// SelectHandler handles slack selector payloads
func (s Slack) SelectHandler(w http.ResponseWriter, r *http.Request) {
	i := &slack.InteractionCallback{}
	if err := json.Unmarshal([]byte(r.FormValue("payload")), i); err != nil {
		log.WithError(err).Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"filter": i.Value,
		"action": i.ActionID,
	}).Debug("selector search")

	resp := slack.OptionsResponse{}
	switch i.ActionID {
	case "repository":
		for _, r := range s.Storage.Repositories.Search(i.Value, 20) {
			resp.Options = append(resp.Options, slack.NewOptionBlockObject(fmt.Sprintf("%d/%s", r.Rank, r.Key()), slack.NewTextBlockObject("plain_text", fmt.Sprintf(":%s: %s", r.ProviderType, r.Name), true, false), nil))
		}
	case "from_ref", "to_ref":
		repoKey := providers.RepositoryKey(i.View.PrivateMetadata)
		repo, found := s.Storage.Repositories.GetByKey(repoKey)
		if !found {
			log.WithField("repository_key", repoKey).WithError(fmt.Errorf("repository not found")).Error()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var err error
		if repo.LastRefsUpdate.Add(2*time.Minute).Unix() < time.Now().Unix() || len(repo.Refs) == 0 {
			repo.Refs, err = s.Providers[repo.ProviderType].ListRefs(repo.Name)
			if err != nil {
				log.WithField("repository_key", repoKey).WithError(fmt.Errorf("could not update repository's refs")).Error()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			repo.LastRefsUpdate = time.Now()
		}

		for _, r := range repo.Refs.Search(i.Value, 20) {
			resp.Options = append(resp.Options, slack.NewOptionBlockObject(fmt.Sprintf("%d/%s", r.Rank, r.Key()), slack.NewTextBlockObject("plain_text", fmt.Sprintf("%s/%s", r.Type, r.Name), true, false), nil))
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
