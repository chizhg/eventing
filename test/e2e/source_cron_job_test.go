// +build e2e

/*
Copyright 2019 The Knative Authors
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
	"fmt"
	"testing"

	"github.com/knative/eventing/test/base"
	"github.com/knative/eventing/test/common"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func TestCronJobSource(t *testing.T) {
	const (
		cronJobSourceName = "e2e-cron-job-source"
		// Every 10 seconds starting at :00 second after the minute
		schedule = "0/10 * * * * ? *"

		loggerPodName = "e2e-cron-job-source-logger-pod"
		saIngressName = "eventing-broker-ingress"
		crIngressName = "eventing-broker-ingress"
	)

	client := Setup(t, true)
	defer TearDown(client)

	// creates ServiceAccount and ClusterRoleBinding with default cluster-admin role
	client.CreateServiceAccountAndBindingOrFail(saIngressName, crIngressName)

	// create event logger pod and service
	loggerPod := base.EventLoggerPod(loggerPodName)
	client.CreatePodOrFail(loggerPod, common.WithService(loggerPodName))

	// create cron job source
	data := fmt.Sprintf("TestCronJobSource %s", uuid.NewUUID())
	sinkOption := base.WithSinkServiceForCronJobSource(loggerPodName)
	client.CreateCronJobSourceOrFail(cronJobSourceName, schedule, data, saIngressName, sinkOption)

	// wait for all test resources to be ready, so that we can start sending events
	if err := client.WaitForAllTestResourcesReady(); err != nil {
		t.Fatalf("Failed to get all test resources ready: %v", err)
	}

	// verify the logger service receives the event
	if err := client.CheckLog(loggerPodName, common.CheckerContainsAtLeast(data, 2)); err != nil {
		t.Fatalf("String %q not found in logs of logger pod %q: %v", data, loggerPodName, err)
	}
}
