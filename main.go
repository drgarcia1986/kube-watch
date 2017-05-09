package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/drgarcia1986/slacker/slack"
)

type Pod struct {
	namespace string
	name      string
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
			checkPods(slackClient)
		}
	}
}

func checkPods(sc *slack.Client) {
	kubectlCmd := getCmd("kubectl", "get", "pods", "--all-namespaces")
	grepCmd := getCmd("grep", "-i", "crash")

	out, err := pipeCmds(kubectlCmd, grepCmd)
	if err != nil {
		fmt.Println("error: ", err)
	}
	pods := cleanOutput(out)
	if len(pods) == 0 {
		return
	}

	msg := fmt.Sprintf(":shit: *PODS IN CRASH* on _%s_:\n\n", k8sEnv)
	for _, pod := range pods {
		msg = fmt.Sprintf("%sNamespace: *%s*\nPod: *%s*\nStatus: *%s*\n\n", msg, pod.namespace, pod.name, pod.status)
	}

	if err = sc.PostMessage(slackChannel, "kube-watch", slackAvatar, msg); err != nil {
		fmt.Println("Error on post msg on slack: ", err)
	}
}

func cleanOutput(out []byte) []Pod {
	pods := make([]Pod, 0)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		s := strings.Fields(line)
		if len(s) == 0 {
			continue
		}
		pods = append(pods, Pod{namespace: s[0], name: s[1], status: s[3]})
	}
	return pods
}

func getCmd(program string, args ...string) *exec.Cmd {
	return exec.Command(program, args...)
}

func pipeCmds(c1, c2 *exec.Cmd) ([]byte, error) {
	in, err := c1.StdoutPipe()
	if err != nil {
		return nil, err
	}
	c2.Stdin = in

	var out bytes.Buffer
	c2.Stdout = &out

	if err := c2.Start(); err != nil {
		return nil, err
	}

	if err := c1.Run(); err != nil {
		return nil, err
	}

	if err := c2.Wait(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
