/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"os"
	"testing"

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
