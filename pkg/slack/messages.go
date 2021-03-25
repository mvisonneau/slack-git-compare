package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/mvisonneau/slack-git-compare/pkg/providers"
	"github.com/openlyinc/pointy"
	"github.com/slack-go/slack"
	"github.com/xeonx/timeago"
)

// ModalRequestOptions ..
type ModalRequestOptions struct {
	ConversationID                  string
	Repository                      providers.Repository
	FromRef                         providers.Ref
	ToRef                           providers.Ref
	Comparison                      *providers.Comparison
	LastRepositoriesUpdate          time.Time
	CurrentlyUpdatingRepositories   bool
	CurrentlyUpdatingRepositoryRefs bool
}

// ViewSubmissionResponse ..
type ViewSubmissionResponse struct {
	ResponseType string            `json:"response_type"`
	Errors       map[string]string `json:"errors,omitempty"`
}

// GetModalRequest ..
func GetModalRequest(opts ModalRequestOptions) (mvr slack.ModalViewRequest) {
	mvr.Type = slack.ViewType("modal")
	mvr.Title = slack.NewTextBlockObject(slack.PlainTextType, "git compare", false, false)
	mvr.Close = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	mvr.Submit = slack.NewTextBlockObject(slack.PlainTextType, "Post to channel", false, false)
	mvr.CallbackID = opts.ConversationID

	if opts.CurrentlyUpdatingRepositories {
		mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, slack.NewSectionBlock(slack.NewTextBlockObject(slack.PlainTextType, ":repeat: updating repostories list..", true, false), nil, nil))
		return
	}

	repositoriesElement := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, "repository")
	repositoriesElement.MinQueryLength = pointy.Int(0)
	repositoriesInput := slack.NewInputBlock(
		"repositories",
		slack.NewTextBlockObject(slack.PlainTextType, "Select a repository", false, false),
		repositoriesElement,
	)
	repositoriesInput.DispatchAction = true

	repositoriesUpdateSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("_updated %s_", timeago.English.Format(opts.LastRepositoriesUpdate)), false, false),
		nil,
		slack.NewAccessory(
			slack.NewButtonBlockElement(
				"update_repositories",
				"",
				slack.NewTextBlockObject(
					slack.PlainTextType,
					"update now",
					false,
					false,
				),
			),
		),
	)

	mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, repositoriesInput)
	mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, repositoriesUpdateSection)

	if !opts.Repository.IsEmpty() {
		// This is only useful when specifying the repository name as part of the slash command
		repositoriesElement.InitialOption = slack.NewOptionBlockObject(
			fmt.Sprintf("x/%s", opts.Repository.Key()),
			slack.NewTextBlockObject(
				slack.PlainTextType,
				fmt.Sprintf(":%s: %s", opts.Repository.ProviderType, opts.Repository.Name),
				true,
				false,
			),
			nil,
		)

		// Add a divider
		mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, slack.NewDividerBlock())

		if !opts.Repository.RefsLastUpdate.IsZero() {
			// FROM REF
			fromRefElement := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, fmt.Sprintf("from_ref/%s", string(opts.Repository.Key())))
			fromRefElement.MinQueryLength = pointy.Int(0)
			if !opts.FromRef.IsEmpty() {
				fromRefElement.InitialOption = slack.NewOptionBlockObject(
					fmt.Sprintf("x/%s", opts.FromRef.Key()),
					slack.NewTextBlockObject(
						slack.PlainTextType,
						fmt.Sprintf("%s/%s", opts.FromRef.Type, opts.FromRef.Name),
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
			toRefElement := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, fmt.Sprintf("to_ref/%s", string(opts.Repository.Key())))
			toRefElement.MinQueryLength = pointy.Int(0)
			if !opts.ToRef.IsEmpty() {
				toRefElement.InitialOption = slack.NewOptionBlockObject(
					fmt.Sprintf("x/%s", opts.ToRef.Key()),
					slack.NewTextBlockObject(
						slack.PlainTextType,
						fmt.Sprintf("%s/%s", opts.ToRef.Type, opts.ToRef.Name),
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

			// Update refs
			refsUpdateSection := slack.NewSectionBlock(
				slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("_updated %s_", timeago.English.Format(opts.Repository.RefsLastUpdate)), false, false),
				nil,
				slack.NewAccessory(
					slack.NewButtonBlockElement(
						"update_refs",
						string(opts.Repository.Key()),
						slack.NewTextBlockObject(
							slack.PlainTextType,
							"update now",
							false,
							false,
						),
					),
				),
			)

			// Add the refs selectors
			mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, fromRefInput)
			mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, toRefInput)
			mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, refsUpdateSection)

			if opts.Comparison != nil {
				// Add a divider
				mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, slack.NewDividerBlock())

				var msg string
				if opts.Comparison.CommitCount() == 0 {
					msg = ":shrug: there are no difference between the refs"
				} else {
					commitString := "commit"
					if opts.Comparison.CommitCount() > 1 {
						commitString += "s"
					}
					msg = fmt.Sprintf("*%d %s* found the between the refs\n%s", opts.Comparison.CommitCount(), commitString, strings.TrimPrefix(opts.Comparison.AuthorsSlackString(), commitString+" "))
				}

				headerButtonText := fmt.Sprintf("View in %s", opts.Repository.ProviderType.StringPretty())
				headerButton := slack.NewButtonBlockElement("", "", slack.NewTextBlockObject("plain_text", headerButtonText, false, false))
				headerButton.URL = opts.Comparison.WebURL

				mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", msg, false, false), nil, slack.NewAccessory(headerButton)))
			}
		} else {
			mvr.Blocks.BlockSet = append(mvr.Blocks.BlockSet, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", ":repeat: updating refs list..", false, false), nil, nil))
		}
	}

	return
}

// GenerateComparisonMessage ..
func GenerateComparisonMessage(repo providers.Repository, fromRef, toRef providers.Ref, cmp providers.Comparison, slackUserID string) slack.Blocks {
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

	if len(cmp.Commits) > 15 {
		commitsText += fmt.Sprintf(":warning: there are *%d commits* between these refs, truncated to the *15* most recent ones\n", len(cmp.Commits))
	}

	for i, c := range cmp.Commits {
		if i >= 15 {
			break
		}

		commitsText += fmt.Sprintf("> <%s|%s> | _%s_\n", c.WebURL, c.ShortID, c.ShortMessage())
	}

	footerText := fmt.Sprintf("diff requested by <@%s> | %s", slackUserID, cmp.AuthorsSlackString())
	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", headerText, false, false), nil, slack.NewAccessory(headerButton)),
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", commitsText, false, false), nil, nil),
			slack.NewDividerBlock(),
			slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", footerText, false, false)),
		},
	}

	return blocks
}
