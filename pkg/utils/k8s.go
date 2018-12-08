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

package utils

import (
	"context"
	"github.com/cedrickring/kbuild/pkg/log"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" //auth for GKE clusters
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
)

//GetClient creates a new kubernetes client with the kubeconfig at ~/.kube/config
func GetClient() (*kubernetes.Clientset, error) {
	home := homeDir()
	if home == "" {
		return nil, errors.New("Can't find kubeconfig at ~/.kube/config")
	}
	kubeconfigPath := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "build client config from flags")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "new clientset from config")
	}

	return clientset, nil
}

//WaitForPodInitialized waits for a specific pod to be initialized
func WaitForPodInitialized(clientset *kubernetes.Clientset, namespace, podName string) error {
	log.Infof("Waiting for pod %s to be initialized", podName)

	pods := clientset.CoreV1().Pods(namespace)

	ctx, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelTimeout()

	return wait.PollImmediateUntil(500*time.Millisecond, func() (done bool, err error) {
		pod, err := pods.Get(podName, metav1.GetOptions{
			IncludeUninitialized: true,
		})

		if err != nil {
			log.Infof("Getting pod %s", podName)
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
func WaitForPodComplete(clientset *kubernetes.Clientset, namespace, podName string) error {
	pods := clientset.CoreV1().Pods(namespace)

	return wait.PollImmediateUntil(500*time.Millisecond, func() (done bool, err error) {
		pod, err := pods.Get(podName, metav1.GetOptions{
			IncludeUninitialized: true,
		})

		if err != nil {
			log.Infof("Getting pod %s", podName)
			return false, nil
		}

		for _, init := range pod.Status.ContainerStatuses {
			if init.State.Terminated != nil {
				return true, nil
			}
		}

		return false, nil
	}, context.Background().Done())
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" { //unix
		return h
	}
	return os.Getenv("USERPROFILE") //windows
}
