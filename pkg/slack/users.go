package slack

import (
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// ListSlackUserEmailMappings returns a mapping between email addresses and slack user IDs
func (s *Slack) ListSlackUserEmailMappings() (mapping map[string]string, err error) {
	mapping = make(map[string]string)
	var users []slack.User
	users, err = s.Client.GetUsers()
	if err != nil {
		return
	}

	for _, u := range users {
		if u.Profile.Email == "" {
			log.WithFields(log.Fields{
				"slack_user_id":   u.ID,
				"slack_user_name": u.Name,
			}).Debug("no email defined for slack user")
			continue
		}

		log.WithFields(log.Fields{
			"slack_user_id":    u.ID,
			"slack_user_name":  u.Name,
			"slack_user_email": u.Profile.Email,
		}).Debug("found new slack user")
		mapping[u.Profile.Email] = u.ID
	}

	for _, customUser := range s.CustomUsers {
		if userID, found := mapping[customUser.Email]; found {
			for _, alias := range customUser.Aliases {
				log.WithFields(log.Fields{
					"slack_user_id":    userID,
					"slack_user_email": customUser.Email,
					"alias":            alias,
				}).Debug("added new email alias for slack user")
				mapping[alias] = userID
			}
		} else {
			log.WithField("email", customUser.Email).Warning("custom user mapping not satisfied, slack user not matched")
		}
	}

	return
}
