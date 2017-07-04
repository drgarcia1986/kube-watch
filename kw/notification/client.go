package notification

import "github.com/drgarcia1986/slacker/slack"

type Notification interface {
	PostMessage(msg string) error
}

type Config struct {
	SlackAvatar  string
	SlackToken   string
	SlackChannel string
}

type Default struct {
	sc     *slack.Client
	config *Config
}

func (d *Default) PostMessage(msg string) error {
	return d.sc.PostMessage(
		d.config.SlackChannel,
		"kube-watch",
		d.config.SlackAvatar,
		msg,
	)
}

func New(config *Config) Notification {
	slackClient := slack.New(config.SlackToken)
	return &Default{
		sc:     slackClient,
		config: config,
	}
}
