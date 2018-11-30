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
	"github.com/cedrickring/kbuild/pkg/docker"
	"github.com/cedrickring/kbuild/pkg/log"
	"github.com/cedrickring/kbuild/pkg/utils"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//Build contains all required information to start a Kaniko build
type Build struct {
	ImageTag       string
	WorkDir        string
	DockerfilePath string
	Cache          bool
	CacheRepo      string

	tarPath string
}

//StartBuild starts a Kaniko build with options provided in `Build`
func (b Build) StartBuild() error {
	client, err := utils.GetClient()
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

	pods := client.CoreV1().Pods("default")
	pod, err := pods.Create(b.getKanikoPod())
	if err != nil {
		return errors.Wrap(err, "creating kaniko pod")
	}
	defer func() {
		log.Info("Deleting build pod...")
		err := pods.Delete(pod.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
		if err != nil {
			log.Err(err)
		}
	}()

	err = b.copyTarIntoPod(client, pod)
	if err != nil {
		return errors.Wrap(err, "copying context into pod")
	}

	log.Info("Starting build...")
	cancel := b.streamLogs(client, pod.Name)

	if err := utils.WaitForPodComplete(client, pod.Name); err != nil {
		return errors.Wrap(err, "waiting for kaniko pod to complete")
	}

	cancel() //stop streaming logs

	podStatus, err := pods.Get(pod.Name, metav1.GetOptions{})
	if err == nil && podStatus.Status.ContainerStatuses[0].State.Terminated.Reason == "Error" { //build container exited with a non 0 code
		return errors.New("Build failed.")
	}

	log.Info("Build succeeded.")
	return nil
}

func (b Build) checkForConfigMap(client *kubernetes.Clientset) error {
	configMap, err := docker.GetConfigAsConfigMap()
	if err != nil {
		return errors.Wrap(err, "get docker config as ConfigMap")
	}

	configMaps := client.CoreV1().ConfigMaps("default")

	_, err = configMaps.Get(configMap.Name, metav1.GetOptions{})
	if err != nil { //configmap is not present
		_, err = configMaps.Create(configMap) //so we create a new one
		if err != nil {
			return errors.Wrap(err, "creating configmap")
		}
	} else {
		_, err := configMaps.Update(configMap) //otherwise update the existing configmap
		if err != nil {
			return errors.Wrap(err, "updating configmap")
		}
	}

	return nil
}

func (b *Build) generateContext() (func(), error) {
	b.tarPath = filepath.Join(os.TempDir(), fmt.Sprintf("context-%s.tar.gz", utils.RandomID()))

	file, err := os.Create(b.tarPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = docker.GetContextFromDir(b.WorkDir, file)
	if err != nil {
		return nil, errors.Wrap(err, "generating context")
	}

	return func() {
		if err := os.Remove(b.tarPath); err != nil {
			log.Err(err)
		}
	}, nil
}

func (b Build) copyTarIntoPod(clientset *kubernetes.Clientset, generatedPod *v1.Pod) error {
	if err := utils.WaitForPodInitialized(clientset, generatedPod.Name); err != nil {
		return errors.Wrap(err, "wait for generatedPod initialized")
	}

	podTarPath := fmt.Sprintf("/tmp/%s", filepath.Base(b.tarPath))
	localTarPath := b.tarPath
	if runtime.GOOS == "windows" {
		localTarPath = localTarPath[2:]
	}
	initContainerName := generatedPod.Spec.InitContainers[0].Name

	log.Info("Copying build context into container...")

	//kubectl cp <tar path> <generatedPod>:<path> -c <initcontainer> [-n <namespace>]
	cp := exec.Command("kubectl", "cp", localTarPath, fmt.Sprintf("%s:%s", generatedPod.Name, podTarPath), "-c", initContainerName)
	if err := cp.Run(); err != nil {
		return errors.Wrap(err, "copying tar into init container")
	}

	//kubectl exec <generatedPod> -c <initcontainer> [-n <namespace>] -- tar -zxf /tmp/<tar> -C /kaniko/build-context
	tar := exec.Command("kubectl", "exec", generatedPod.Name, "-c", initContainerName, "--", "tar", "-zxf", podTarPath, "-C", "/kaniko/build-context")
	if err := tar.Run(); err != nil {
		return errors.Wrap(err, "extracting tar in init container")
	}

	//kubectl exec <generatedPod> -c <initcontainer> [-n <namespace>] -- touch /tmp/complete
	touch := exec.Command("kubectl", "exec", generatedPod.Name, "-c", initContainerName, "--", "touch", "/tmp/complete")
	if err := touch.Run(); err != nil {
		return errors.Wrap(err, "creating complete file in init container")
	}

	log.Info("Finished copying build context.")
	return nil
}
