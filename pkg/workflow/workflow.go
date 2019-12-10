// Copyright 2019 The SQLFlow Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"fmt"
	"os/exec"
	"regexp"
	"time"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	pb "sqlflow.org/sqlflow/pkg/proto"
)

func isCompletedPhase(phase wfv1.NodePhase) bool {
	return phase == wfv1.NodeSucceeded ||
		phase == wfv1.NodeFailed ||
		phase == wfv1.NodeError ||
		phase == wfv1.NodeSkipped
}

func getWorkflowID(output string) (string, error) {
	reWorkflow := regexp.MustCompile(`.+/(.+) .+`)
	wf := reWorkflow.FindStringSubmatch(string(output))
	if len(wf) != 2 {
		return "", fmt.Errorf("parse workflow ID error: %v", output)
	}

	return wf[1], nil
}

func getWorkflowResource(job pb.Job) (*wfv1.Workflow, error) {
	cmd := exec.Command("kubectl", "get", "wf", job.Id, "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("getWorkflowResource error: %v\n%v", string(output), err)
	}
	return parseWorkflowResource(output)
}

func getWorkflowStatusPhase(job pb.Job) (wfv1.NodePhase, error) {
	wf, err := getWorkflowResource(job)
	if err != nil {
		return "", fmt.Errorf("getWorkflowStatusPhase error: %v", err)
	}
	return wf.Status.Phase, nil
}

func getWorkflowPodName(job pb.Job) (string, error) {
	cmd := exec.Command("kubectl", "get", "pods",
		fmt.Sprintf(`--selector=workflows.argoproj.io/workflow=%s`, job.Id),
		"-o", "jsonpath={.items[0].metadata.name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getWorkflowPodName error: %v\n%v", string(output), err)
	}

	return string(output), nil
}

func getPodLogs(podName string) (string, error) {
	// NOTE(tony): A workflow pod usually contains two container: main and wait
	// I believe wait is used for management by Argo, so we only need to care about main.
	cmd := exec.Command("kubectl", "logs", podName, "main")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getPodLogs error: %v\n%v", string(output), err)
	}
	return string(output), nil
}

func fetchWorkflowLog(job pb.Job) (string, error) {
	for {
		statusPhase, err := getWorkflowStatusPhase(job)
		if err != nil {
			return "", err
		}

		// FIXME(tony): what if it is a long running job
		if isCompletedPhase(statusPhase) {
			break
		}
		time.Sleep(time.Second)
	}

	// FIXME(tony): what if there are multiple pods
	podName, err := getWorkflowPodName(job)
	if err != nil {
		return "", err
	}

	return getPodLogs(podName)
}

// Submit the Argo workflow and returns the workflow ID
func Submit(argoFileName string) (string, error) {
	// submit Argo YAML and fetch the workflow ID.
	cmd := exec.Command("kubectl", "create", "-f", argoFileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("submit Argo YAML error: %v, output: %s", err, string(output))
	}

	workflowID, err := getWorkflowID(string(output))
	if err != nil {
		return "", err
	}
	return workflowID, err
}