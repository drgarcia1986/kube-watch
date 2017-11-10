package kw

import (
	"fmt"
	"time"

	"github.com/drgarcia1986/kube-watch/kw/k8s"
	"github.com/drgarcia1986/kube-watch/kw/notification"
)

type KubeWatch struct {
	k                 k8s.Client
	n                 notification.Notification
	ct                int
	notReadyThreshold int
	k8sEnv            string
}

func (kw *KubeWatch) Run() {
	for {
		select {
		case <-time.After(time.Duration(kw.ct) * time.Minute):
			kw.checkCrashedPods()
			kw.checkNotReadyPods()
		}
	}
}

func (kw *KubeWatch) checkCrashedPods() {
	podList, err := kw.k.List("")
	if err != nil {
		kw.propagateMsg(fmt.Sprintf(":bomb: Error on check pods status: *%v*", err))
		return
	}

	podsInCrash := groupByNamespace(podList)
	filterCrashedsPods(podsInCrash)
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

func (kw *KubeWatch) checkNotReadyPods() {
	podList, err := kw.k.List("")
	if err != nil {
		kw.propagateMsg(fmt.Sprintf(":bomb: Error on check pods status: *%v*", err))
		return
	}

	podsByNamespace := groupByNamespace(podList)
	namespaceWithNotReadyPods := podsNotReadyByThreshold(podsByNamespace, kw.notReadyThreshold)
	if len(namespaceWithNotReadyPods) == 0 {
		return
	}

	msg := fmt.Sprintf(":warning: *NAMESPACES WITH HIGH NUMBER OF PODS NOT READY* on _%s_:\n\n", kw.k8sEnv)
	for ns, perc := range namespaceWithNotReadyPods {
		team, err := kw.k.GetLabelValue(ns, "teresa.io/team")
		if err != nil {
			kw.propagateMsg(fmt.Sprintf(":bomb: Error getting namespace label: *%v*", err))
		}
		msg = fmt.Sprintf("%s*%s*: (@%s) *%d %%* of pods Not Ready\n", msg, ns, team, perc)
	}

	if err = kw.propagateMsg(msg); err != nil {
		fmt.Println("Error on post msg on slack: ", err)
	}
}

func podsNotReadyByThreshold(items map[string][]k8s.Pod, threshold int) map[string]int {
	result := make(map[string]int)
	for ns, pods := range items {
		count := 0
		for _, pod := range pods {
			if pod.Status == "Running" && !pod.Ready {
				count++
			}
		}
		percNotReady := (count * 100) / len(pods)
		if percNotReady >= threshold {
			result[ns] = percNotReady
		}
	}
	return result
}

func (kw *KubeWatch) propagateMsg(msg string) error {
	fmt.Println(msg)
	return kw.n.PostMessage(msg)
}

func filterCrashedsPods(items map[string][]k8s.Pod) {
	for ns, pods := range items {
		podsInCrash := make([]k8s.Pod, 0)

		for _, pod := range pods {
			if pod.Status == "CrashLoopBackOff" {
				podsInCrash = append(podsInCrash, pod)
			}
		}

		if len(podsInCrash) > 0 {
			items[ns] = podsInCrash
		} else {
			delete(items, ns)
		}
	}
}

func groupByNamespace(items []k8s.Pod) map[string][]k8s.Pod {
	result := make(map[string][]k8s.Pod)
	for _, pod := range items {
		if _, found := result[pod.Namespace]; !found {
			result[pod.Namespace] = make([]k8s.Pod, 0)
		}
		result[pod.Namespace] = append(result[pod.Namespace], pod)
	}
	return result
}

func New(k k8s.Client, n notification.Notification, circleTime, notReadyThreshold int, k8sEnv string) *KubeWatch {
	return &KubeWatch{
		k:                 k,
		n:                 n,
		ct:                circleTime,
		k8sEnv:            k8sEnv,
		notReadyThreshold: notReadyThreshold,
	}
}
