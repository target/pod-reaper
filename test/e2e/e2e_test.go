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

	"github.com/target/pod-reaper/cmd/pod-reaper/app"
)

func TestE2E(t *testing.T) {
	// If we have reached here, it means cluster would have been already setup and the kubeconfig file should
	// be in /tmp directory as admin.conf.
	os.Setenv("RUN_DURATION", "1s")
	os.Setenv("CHAOS_CHANCE", "0.5")
	reaper := app.NewReaper("/tmp/admin.conf")
	reaper.Harvest()
}
