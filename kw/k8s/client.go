package k8s

import (
	"k8s.io/client-go/kubernetes"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type Pod struct {
	Name      string
	Namespace string
	Status    string
	Ready     bool
}

type Client interface {
	List(namespace string) ([]Pod, error)
	GetLabelValue(namespace, label string) (string, error)
}

type Default struct {
	k *kubernetes.Clientset
}

func (d *Default) List(namespace string) ([]Pod, error) {
	podList, err := d.k.CoreV1().Pods("").List(k8sv1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return convertPodList(podList.Items), nil
}

func (d *Default) GetLabelValue(namespace, label string) (string, error) {
	ns, err := d.k.CoreV1().Namespaces().Get(namespace)
	if err != nil {
		return "", nil
	}
	return ns.Labels[label], nil
}

func convertPodList(items []k8sv1.Pod) []Pod {
	pods := make([]Pod, 0)
	for _, pod := range items {
		for _, status := range pod.Status.ContainerStatuses {
			state := "Running"
			if status.State.Waiting != nil {
				state = status.State.Waiting.Reason
			} else if status.State.Terminated != nil {
				state = status.State.Terminated.Reason
			}
			pods = append(pods, Pod{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Status:    state,
				Ready:     status.Ready})
		}
	}
	return pods
}

func newK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func New() (Client, error) {
	k, err := newK8sClient()
	if err != nil {
		return nil, err
	}
	return &Default{k: k}, nil
}
