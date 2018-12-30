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

package kaniko

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/cedrickring/kbuild/pkg/constants"
	"github.com/cedrickring/kbuild/pkg/docker"
	"github.com/cedrickring/kbuild/pkg/kubernetes"
	"github.com/cedrickring/kbuild/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"os"
	"path/filepath"
)

//Build contains all required information to start a Kaniko build
type Build struct {
	ImageTags      []string
	WorkDir        string
	DockerfilePath string
	Cache          bool
	CacheRepo      string
	Namespace      string
	BuildArgs      []string
	CredentialsMap *v1.ConfigMap

	tarPath string
}

//ErrorBuildFailed is an error for a failed build
var ErrorBuildFailed = errors.New("build failed")

//StartBuild starts a Kaniko build with options provided in `Build`
func (b Build) StartBuild() error {
	client, err := kubernetes.GetClient()
	if err != nil {
		return errors.Wrap(err, "get kubernetes client")
	}

	err = b.checkForConfigMap(client)
	if err != nil {
		return errors.Wrap(err, "check for config map")
	}

	cleanup, err := b.generateContext()
	if err != nil {
		return err
	}
	defer cleanup()

	pods := client.CoreV1().Pods(b.Namespace)
	pod, err := pods.Create(b.getKanikoPod())
	if err != nil {
		return errors.Wrap(err, "creating kaniko pod")
	}
	defer func() {
		logrus.Info("Deleting build pod...")
		err := pods.Delete(pod.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			logrus.Error(err)
		}
	}()

	err = b.copyTarIntoPod(client, pod)
	if err != nil {
		return errors.Wrap(err, "copying context into pod")
	}

	logrus.Info("Starting build...")
	cancel := b.streamLogs(client, pod.Name)

	if err := kubernetes.WaitForPodComplete(client, b.Namespace, pod.Name); err != nil {
		return errors.Wrap(err, "waiting for kaniko pod to complete")
	}

	cancel() //stop streaming logs

	podStatus, err := pods.Get(pod.Name, metav1.GetOptions{})
	if err == nil && podStatus.Status.ContainerStatuses[0].State.Terminated.Reason == "Error" { //build container exited with a non 0 code
		return ErrorBuildFailed
	}

	logrus.Info("Build succeeded.")
	return nil
}

func (b Build) checkForConfigMap(client *k8s.Clientset) error {
	configMaps := client.CoreV1().ConfigMaps(b.Namespace)

	_, err := configMaps.Get(b.CredentialsMap.Name, metav1.GetOptions{})
	if err != nil { //configmap is not present
		_, err = configMaps.Create(b.CredentialsMap) //so we create a new one
		if err != nil {
			return errors.Wrap(err, "creating configmap")
		}
	} else {
		_, err := configMaps.Update(b.CredentialsMap) //otherwise update the existing configmap
		if err != nil {
			return errors.Wrap(err, "updating configmap")
		}
	}

	return nil
}

func (b *Build) generateContext() (func(), error) {
	b.tarPath = filepath.Join(os.TempDir(), fmt.Sprintf("context-%s.tar.gz", util.RandomID()))

	file, err := os.Create(b.tarPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = docker.CreateContextFromWorkingDir(b.WorkDir, b.DockerfilePath, file, b.BuildArgs)
	if err != nil {
		return nil, errors.Wrap(err, "generating context")
	}

	return func() {
		if err := os.Remove(b.tarPath); err != nil {
			logrus.Error(err)
		}
	}, nil
}

func (b Build) copyTarIntoPod(clientset *k8s.Clientset, generatedPod *v1.Pod) error {
	if err := kubernetes.WaitForPodInitialized(clientset, b.Namespace, generatedPod.Name); err != nil {
		return errors.Wrap(err, "wait for generatedPod initialized")
	}

	logrus.Info("Copying build context into container...")
	initContainerName := generatedPod.Spec.InitContainers[0].Name

	tarCopy := kubernetes.Copy{
		Namespace: b.Namespace,
		PodName:   generatedPod.Name,
		Container: initContainerName,
		SrcPath:   b.tarPath,
		DestPath:  constants.KanikoBuildContextPath,
	}
	if err := tarCopy.CopyFileIntoPod(clientset); err != nil {
		return errors.Wrap(err, "copying tar into init container")
	}

	touch := kubernetes.Exec{
		Namespace: b.Namespace,
		PodName:   generatedPod.Name,
		Container: initContainerName,
		Command:   []string{"touch", "/tmp/complete"},
	}
	if err := touch.Exec(clientset); err != nil {
		return errors.Wrap(err, "creating complete file in init container")
	}

	logrus.Info("Finished copying build context.")
	return nil
}
