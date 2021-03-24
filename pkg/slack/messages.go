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

// GenerateModalRequestRepositoryPicker ..
func GenerateModalRequestRepositoryPicker(
	conversationID string,
	repo providers.Repository,
	lastRepoUpdate time.Time,
	fromRef, toRef providers.Ref,
	lastRefsUpdate time.Time,
	cmp *providers.Comparison,
) slack.ModalViewRequest {

	modalRequest := slack.ModalViewRequest{}
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = slack.NewTextBlockObject(slack.PlainTextType, "git compare", false, false)
	modalRequest.Close = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	modalRequest.Submit = slack.NewTextBlockObject(slack.PlainTextType, "Post to channel", false, false)
	modalRequest.CallbackID = conversationID

	repositoriesElement := slack.NewOptionsSelectBlockElement(slack.OptTypeExternal, nil, "repository")
	repositoriesElement.MinQueryLength = pointy.Int(0)
	repositoriesInput := slack.NewInputBlock(
		"repositories",
		slack.NewTextBlockObject(slack.PlainTextType, "Select a repository", false, false),
		repositoriesElement,
	)
	repositoriesInput.DispatchAction = true

	repositoriesUpdateSection := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("_updated %s_", timeago.English.Format(lastRepoUpdate)), false, false),
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

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			repositoriesInput,
			repositoriesUpdateSection,
		},
	}

	if !repo.IsEmpty() {
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
		if !fromRef.IsEmpty() {
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
		if !toRef.IsEmpty() {
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

		// Update refs
		refsUpdateSection := slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("_updated %s_", timeago.English.Format(lastRefsUpdate)), false, false),
			nil,
			slack.NewAccessory(
				slack.NewButtonBlockElement(
					"update_refs",
					string(repo.Key()),
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
		blocks.BlockSet = append(blocks.BlockSet, fromRefInput)
		blocks.BlockSet = append(blocks.BlockSet, toRefInput)
		blocks.BlockSet = append(blocks.BlockSet, refsUpdateSection)

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
