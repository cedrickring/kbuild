/*
   Copyright 2018 Cedric Kring

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

package kubernetes

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

//WaitForPodInitialized waits for a specific pod to be initialized
func WaitForPodInitialized(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string) error {
	logrus.Infof("Waiting for pod %s to be initialized", podName)

	pods := clientset.CoreV1().Pods(namespace)

	ctx, cancelTimeout := context.WithTimeout(ctx, 10*time.Minute)
	defer cancelTimeout()

	return wait.PollImmediateUntil(500*time.Millisecond, func() (done bool, err error) {
		pod, err := pods.Get(podName, metav1.GetOptions{})

		if err != nil {
			logrus.Infof("Getting pod %s", podName)
			return false, nil
		}

		for _, init := range pod.Status.InitContainerStatuses {
			if init.State.Running != nil {
				return true, nil
			}
		}

		return false, nil
	}, ctx.Done())
}

//WaitForPodComplete waits for a specific pod to be in complete state
func WaitForPodComplete(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string, finish chan bool) error {
	pods := clientset.CoreV1().Pods(namespace)

	return wait.PollImmediateUntil(500*time.Millisecond, func() (done bool, err error) {
		pod, err := pods.Get(podName, metav1.GetOptions{})

		if err != nil {
			logrus.Infof("Getting pod %s", podName)
			return false, nil
		}

		for _, init := range pod.Status.ContainerStatuses {
			if init.State.Terminated != nil {
				finish <- true
				return true, nil
			}
		}

		return false, nil
	}, ctx.Done())
}
