package client

import (
	"fmt"
	"io/ioutil"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// KubeUtil contains utility functions to interact with a Kubernetes cluster
type KubeUtil struct {
	client  clientset.Interface
	timeout time.Duration
}

const retryInterval = 500 * time.Millisecond
const defaultTimeout = 60 * time.Second

// NewKubeUtil returns a new KubeUtil object that interacts with the given Kubernetes cluster
func NewKubeUtil(client clientset.Interface) KubeUtil {
	return KubeUtil{
		client:  client,
		timeout: defaultTimeout,
	}
}

// ApplyPodManifest applies manifest file to specified namespace
func (k *KubeUtil) ApplyPodManifest(namespace string, path string) (*v1.Pod, error) {
	manifest, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pod *v1.Pod
	err = yaml.Unmarshal(manifest, &pod)
	if err != nil {
		return nil, err
	}

	return k.client.CoreV1().Pods(namespace).Create(pod)
}

// WaitForPodPhase waits until a pod enters the specified phase or timeout is reached
func (k *KubeUtil) WaitForPodPhase(namespace string, name string, status v1.PodPhase) error {
	return wait.PollImmediate(retryInterval, k.timeout, func() (bool, error) {
		getOpts := metav1.GetOptions{}
		pod, err := k.client.CoreV1().Pods(namespace).Get(name, getOpts)
		if err != nil {
			return false, fmt.Errorf("error getting pod name %s: %v", name, err)
		}
		return pod.Status.Phase == status, nil
	})
}

// WaitForPodToDie waits until pod no longer exists or timeout is reached
func (k *KubeUtil) WaitForPodToDie(namespace string, name string) error {
	return wait.PollImmediate(retryInterval, k.timeout, func() (bool, error) {
		_, err := k.client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err == nil {
			return false, nil
		}
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
}
