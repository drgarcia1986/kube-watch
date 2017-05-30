package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"

	"github.com/drgarcia1986/slacker/slack"
)

type Pod struct {
	name      string
	namespace string
	status    string
}

const defaultTimeCircle = 5

var (
	k8sEnv       string = os.Getenv("K8SENV")
	slackAvatar  string = os.Getenv("SLACKAVATAR")
	slackToken   string = os.Getenv("SLACKTOKEN")
	slackChannel string = os.Getenv("SLACKCHANNEL")
	circleTime   string = os.Getenv("CIRCLETIME")
)

func main() {
	k8sClient, err := newK8sClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error on get k8s client:", err)
		return
	}

	slackClient := slack.New(slackToken)
	ct, err := strconv.Atoi(circleTime)
	if err != nil {
		fmt.Printf("Using default time circle: %d\n", defaultTimeCircle)
		ct = defaultTimeCircle
	}

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-quit:
			return
		case <-time.After(time.Duration(ct) * time.Minute):
			checkPods(k8sClient, slackClient)
		}
	}
}

func newK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func checkPods(kc *kubernetes.Clientset, sc *slack.Client) {
	podList, err := kc.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		propagateMsg(sc, fmt.Sprintf(":bomb: Error on check pods status: *%v*", err))
		return
	}

	podsInCrash := getPodsInCrash(podList.Items)
	if len(podsInCrash) == 0 {
		return
	}

	msg := fmt.Sprintf(":shit: *PODS IN CRASH* on _%s_:\n\n", k8sEnv)
	for _, pod := range podsInCrash {
		msg = fmt.Sprintf(
			"%sNamespace: *%s*\nPod: *%s*\nStatus: *%s*\n\n",
			msg, pod.namespace, pod.name, pod.status,
		)
	}

	if err = propagateMsg(sc, msg); err != nil {
		fmt.Println("Error on post msg on slack: ", err)
	}
}

func getPodsInCrash(items []k8sv1.Pod) []Pod {
	podsInCrash := make([]Pod, 0)
	for _, pod := range items {
		for _, status := range pod.Status.ContainerStatuses {
			var state string
			if status.State.Waiting != nil {
				state = status.State.Waiting.Reason
			} else if status.State.Terminated != nil {
				state = status.State.Terminated.Reason
			}

			if state == "CrashLoopBackOff" {
				podsInCrash = append(
					podsInCrash,
					Pod{name: pod.Name, namespace: pod.Namespace, status: state},
				)
			}
		}
	}
	return podsInCrash
}

func propagateMsg(sc *slack.Client, msg string) error {
	fmt.Println(msg)
	return sc.PostMessage(slackChannel, "kube-watch", slackAvatar, msg)
}
