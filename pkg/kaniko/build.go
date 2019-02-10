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
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/cedrickring/kbuild/pkg/docker"
	"github.com/cedrickring/kbuild/pkg/kaniko/source"
	"github.com/cedrickring/kbuild/pkg/kubernetes"
	"github.com/cedrickring/kbuild/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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
	Source         source.Source

	tarPath string
}

//ErrorBuildFailed is an error for a failed build
var ErrorBuildFailed = errors.New("build failed")

//StartBuild starts a Kaniko build with options provided in `Build`
func (b Build) StartBuild(ctx context.Context) error {
	client, err := kubernetes.GetClient()
	if err != nil {
		return errors.Wrap(err, "get kubernetes client")
	}

	cleanup, err := b.checkForConfigMap(client)
	if err != nil {
		return errors.Wrap(err, "check for config map")
	}
	defer cleanup()

	cleanup, err = b.generateContext()
	if err != nil {
		return err
	}
	defer cleanup()

	pod := b.getKanikoPod()
	b.Source.ModifyPod(pod)

	if err := b.Source.PrepareCredentials(); err != nil {
		return errors.Wrap(err, "preparing credentials")
	}

	if !b.Source.RequiresPod() {
		if err := b.Source.UploadTar(pod, b.tarPath); err != nil {
			return errors.Wrap(err, "uploading tar")
		}
	}

	pods := client.CoreV1().Pods(b.Namespace)
	pod, err = pods.Create(pod)
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

	if b.Source.RequiresPod() {
		if err := b.Source.UploadTar(pod, b.tarPath); err != nil {
			return errors.Wrap(err, "uploading tar")
		}
	}

	defer b.Source.Cleanup()

	logrus.Info("Starting build...")
	cancel := b.streamLogs(ctx, client, pod.Name)

	finishChan := make(chan bool, 1)

	go func() {
		if err := kubernetes.WaitForPodComplete(ctx, client, b.Namespace, pod.Name, finishChan); err != nil && err != wait.ErrWaitTimeout {
			logrus.Error(errors.Wrap(err, "waiting for kaniko pod to complete"))
		}
	}()

	select {
	case <-ctx.Done():
		logrus.Infoln("Build was cancelled")
	case <-finishChan:
		podStatus, err := pods.Get(pod.Name, metav1.GetOptions{})
		if err == nil && podStatus.Status.ContainerStatuses[0].State.Terminated.Reason == "Error" { //build container exited with a non 0 code
			return ErrorBuildFailed
		}

		logrus.Info("Build succeeded.")
	}

	cancel() //stop streaming logs

	return nil
}

func (b Build) checkForConfigMap(client *k8s.Clientset) (func(), error) {
	configMaps := client.CoreV1().ConfigMaps(b.Namespace)

	_, err := configMaps.Get(b.CredentialsMap.Name, metav1.GetOptions{})
	if err != nil { //configmap is not present
		_, err = configMaps.Create(b.CredentialsMap) //so we create a new one
		if err != nil {
			return nil, errors.Wrap(err, "creating configmap")
		}
	} else {
		_, err := configMaps.Update(b.CredentialsMap) //otherwise update the existing configmap
		if err != nil {
			return nil, errors.Wrap(err, "updating configmap")
		}
	}

	return func() {
		logrus.Infoln("Deleting credentials map")
		if err := configMaps.Delete(b.CredentialsMap.Name, &metav1.DeleteOptions{}); err != nil {
			logrus.Error(errors.Wrap(err, "deleting credentials configmap"))
		}
	}, nil
}

func (b *Build) generateContext() (func(), error) {
	b.tarPath = filepath.Join(os.TempDir(), fmt.Sprintf("context-%s.tar.gz", util.RandomID()))

	file, err := os.Create(b.tarPath)
	if err != nil {
		return nil, errors.Wrap(err, "creating tar file")
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
