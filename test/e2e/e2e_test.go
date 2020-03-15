package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/pod-reaper/cmd/pod-reaper/app"
	"github.com/target/pod-reaper/internal/pkg/client"
	v1 "k8s.io/api/core/v1"
)

const namespace string = "default"
const kubeConfig string = "/tmp/admin.conf"

func TestChaosRule(t *testing.T) {
	clientset, err := client.CreateClient(kubeConfig)
	assert.NoError(t, err, "error creating kubernetes client")
	kubeUtil := client.NewKubeUtil(clientset)

	// Create pod and wait for it to be running
	pod, err := kubeUtil.ApplyPodManifest(namespace, "./pause-pod.yml")
	assert.NoError(t, err, "error applying pod manifest")
	kubeUtil.WaitForPodPhase(namespace, pod.Name, v1.PodRunning)

	// Run reaper
	os.Clearenv()
	os.Setenv("NAMESPACE", namespace)
	os.Setenv("RUN_DURATION", "1s")
	os.Setenv("SCHEDULE", "@every 1s")
	os.Setenv("CHAOS_CHANCE", "1")
	reaper := app.NewReaper(clientset)
	reaper.Harvest()

	// Wait for pod to die so other tests aren't affected
	err = kubeUtil.WaitForPodToDie(namespace, pod.Name)
	assert.NoError(t, err, "timed out waiting for pod to die")
}

func TestDurationRule(t *testing.T) {
	clientset, err := client.CreateClient(kubeConfig)
	assert.NoError(t, err, "error creating kubernetes client")
	kubeUtil := client.NewKubeUtil(clientset)

	// Create pod and wait for it to be running
	pod, err := kubeUtil.ApplyPodManifest(namespace, "./pause-pod.yml")
	assert.NoError(t, err, "error applying pod manifest")
	kubeUtil.WaitForPodPhase(namespace, pod.Name, v1.PodRunning)

	// Make sure pod has been running at least 1s
	time.Sleep(1 * time.Second)

	// Run reaper
	os.Clearenv()
	os.Setenv("NAMESPACE", namespace)
	os.Setenv("RUN_DURATION", "1s")
	os.Setenv("SCHEDULE", "@every 1s")
	os.Setenv("MAX_DURATION", "1s")
	reaper := app.NewReaper(clientset)
	reaper.Harvest()

	// Wait for pod to die so other tests aren't affected
	err = kubeUtil.WaitForPodToDie(namespace, pod.Name)
	assert.NoError(t, err, "timed out waiting for pod to die")
}

func TestUnreadyRule(t *testing.T) {
	clientset, err := client.CreateClient(kubeConfig)
	assert.NoError(t, err, "error creating kubernetes client")
	kubeUtil := client.NewKubeUtil(clientset)

	// Create pod and wait for it to be running
	pod, err := kubeUtil.ApplyPodManifest(namespace, "./unready-pod.yml")
	assert.NoError(t, err, "error applying pod manifest")
	kubeUtil.WaitForPodPhase(namespace, pod.Name, v1.PodRunning)

	// Run reaper
	os.Clearenv()
	os.Setenv("NAMESPACE", namespace)
	os.Setenv("RUN_DURATION", "1s")
	os.Setenv("SCHEDULE", "@every 1s")
	os.Setenv("MAX_UNREADY", "1s")
	reaper := app.NewReaper(clientset)
	reaper.Harvest()

	// Wait for pod to die so other tests aren't affected
	err = kubeUtil.WaitForPodToDie(namespace, pod.Name)
	assert.NoError(t, err, "timed out waiting for pod to die")
}
