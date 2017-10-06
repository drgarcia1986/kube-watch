package kw

import (
	"fmt"
	"time"

	"github.com/drgarcia1986/kube-watch/kw/k8s"
	"github.com/drgarcia1986/kube-watch/kw/notification"
)

type KubeWatch struct {
	k      k8s.Client
	n      notification.Notification
	ct     int
	k8sEnv string
}

func (kw *KubeWatch) Run() {
	for {
		select {
		case <-time.After(time.Duration(kw.ct) * time.Minute):
			kw.checkPods()
		}
	}
}

func (kw *KubeWatch) checkPods() {
	podList, err := kw.k.List("")
	if err != nil {
		kw.propagateMsg(fmt.Sprintf(":bomb: Error on check pods status: *%v*", err))
		return
	}

	podsInCrash := groupByNamespaceAndFilterCrasheds(podList)
	if len(podsInCrash) == 0 {
		return
	}

	msg := fmt.Sprintf(":shit: *PODS IN CRASH* on _%s_:\n\n", kw.k8sEnv)
	for ns, pods := range podsInCrash {
		team, err := kw.k.GetLabelValue(ns, "teresa.io/team")
		if err != nil {
			kw.propagateMsg(fmt.Sprintf(":bomb: Error getting namespace label: *%v*", err))
			return
		}
		msg = fmt.Sprintf(
			"%s*%s*: (@%s) *%d* Pod(s) in CrashLoopBackOff\n",
			msg, ns, team, len(pods))
	}

	if err = kw.propagateMsg(msg); err != nil {
		fmt.Println("Error on post msg on slack: ", err)
	}
}

func (kw *KubeWatch) propagateMsg(msg string) error {
	fmt.Println(msg)
	return kw.n.PostMessage(msg)
}

func groupByNamespaceAndFilterCrasheds(items []k8s.Pod) map[string][]k8s.Pod {
	result := make(map[string][]k8s.Pod)
	for _, pod := range items {
		if pod.Status != "CrashLoopBackOff" {
			continue
		}
		if result[pod.Namespace] == nil {
			result[pod.Namespace] = make([]k8s.Pod, 0)
		}
		result[pod.Namespace] = append(result[pod.Namespace], pod)
	}
	return result
}

func New(k k8s.Client, n notification.Notification, circleTime int, k8sEnv string) *KubeWatch {
	return &KubeWatch{
		k:      k,
		n:      n,
		ct:     circleTime,
		k8sEnv: k8sEnv,
	}
}
