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

	podsInCrash := getPodsInCrash(podList)
	if len(podsInCrash) == 0 {
		return
	}

	msg := fmt.Sprintf(":shit: *PODS IN CRASH* on _%s_:\n\n", kw.k8sEnv)
	for _, pod := range podsInCrash {
		msg = fmt.Sprintf(
			"%sNamespace: *%s*\nPod: *%s*\nStatus: *%s*\n\n",
			msg, pod.Namespace, pod.Name, pod.Status,
		)
	}

	if err = kw.propagateMsg(msg); err != nil {
		fmt.Println("Error on post msg on slack: ", err)
	}
}

func (kw *KubeWatch) propagateMsg(msg string) error {
	fmt.Println(msg)
	return kw.n.PostMessage(msg)
}

func getPodsInCrash(items []k8s.Pod) []k8s.Pod {
	podsInCrash := make([]k8s.Pod, 0)
	for _, pod := range items {
		if pod.Status == "CrashLoopBackOff" {
			podsInCrash = append(podsInCrash, pod)
		}
	}
	return podsInCrash
}

func New(k k8s.Client, n notification.Notification, circleTime int, k8sEnv string) *KubeWatch {
	return &KubeWatch{
		k:      k,
		n:      n,
		ct:     circleTime,
		k8sEnv: k8sEnv,
	}
}
