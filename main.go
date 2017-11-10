package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/drgarcia1986/kube-watch/kw"
	"github.com/drgarcia1986/kube-watch/kw/k8s"
	"github.com/drgarcia1986/kube-watch/kw/notification"
)

const (
	defaultTimeCircle        = 5
	defaultNotReadyThreshold = 60
)

var (
	k8sEnv            string = os.Getenv("K8SENV")
	slackAvatar       string = os.Getenv("SLACKAVATAR")
	slackToken        string = os.Getenv("SLACKTOKEN")
	slackChannel      string = os.Getenv("SLACKCHANNEL")
	circleTime        string = os.Getenv("CIRCLETIME")
	notReadyThreshold string = os.Getenv("NOT_READY_THRESHOLD")
)

func main() {
	k8sClient, err := k8s.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error on get k8s client:", err)
		return
	}

	nf := notification.New(&notification.Config{
		SlackAvatar:  slackAvatar,
		SlackToken:   slackToken,
		SlackChannel: slackChannel,
	})

	ct, err := strconv.Atoi(circleTime)
	if err != nil {
		fmt.Printf("Using default time circle: %d\n", defaultTimeCircle)
		ct = defaultTimeCircle
	}

	nrt, err := strconv.Atoi(notReadyThreshold)
	if err != nil {
		fmt.Printf("Using default threshold for not ready pods: %d\n", defaultNotReadyThreshold)
		nrt = defaultNotReadyThreshold
	}

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	kubeWatch := kw.New(k8sClient, nf, ct, nrt, k8sEnv)
	go kubeWatch.Run()

	<-quit
}
