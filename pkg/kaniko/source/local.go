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

package source

import (
	"context"

	"github.com/cedrickring/kbuild/pkg/constants"
	"github.com/cedrickring/kbuild/pkg/kubernetes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

//Local represents a local build context which gets uploaded to an init container
type Local struct {
	Ctx       context.Context
	Namespace string
}

//Cleanup not needed here
func (Local) Cleanup() {
}

//RequiresPod returns always true, since the init pod is required to upload the context
func (Local) RequiresPod() bool {
	return true
}

//PrepareCredentials not needed here
func (l Local) PrepareCredentials() error {
	return nil
}

//ModifyPod adds an init container to the pod and an empty volume for the build context
func (Local) ModifyPod(pod *v1.Pod) {
	//Create init container
	pod.Spec.InitContainers = []v1.Container{
		{
			Name:  "kaniko-init",
			Image: "alpine",
			Args: []string{"sh", "-c",
				`while true; do
							sleep 1; if [ -f /tmp/complete ]; then break; fi
						done`,
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "build-context",
					MountPath: constants.KanikoBuildContextPath,
				},
			},
		},
	}

	//Add dir:// argument
	pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, "--context=dir:///kaniko/build-context")

	//Add volume mount to Kaniko container
	pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
		Name:      "build-context",
		MountPath: constants.KanikoBuildContextPath,
	})

	//Create volume for the build context
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: "build-context",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})
}

//UploadTar uploads the context tar to the init container
func (l Local) UploadTar(pod *v1.Pod, tarPath string) error {
	client, err := kubernetes.GetClient()
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	if err := kubernetes.WaitForPodInitialized(l.Ctx, client, l.Namespace, pod.Name); err != nil && err != wait.ErrWaitTimeout {
		return errors.Wrap(err, "wait for pod initialized")
	}

	logrus.Info("Copying build context into container...")
	initContainerName := pod.Spec.InitContainers[0].Name

	tarCopy := kubernetes.Copy{
		Namespace: l.Namespace,
		PodName:   pod.Name,
		Container: initContainerName,
		SrcPath:   tarPath,
		DestPath:  constants.KanikoBuildContextPath,
	}
	if err := tarCopy.CopyFileIntoPod(client); err != nil {
		return errors.Wrap(err, "copying tar into init container")
	}

	touch := kubernetes.Exec{
		Namespace: l.Namespace,
		PodName:   pod.Name,
		Container: initContainerName,
		Command:   []string{"touch", "/tmp/complete"},
	}
	if err := touch.Exec(client); err != nil {
		return errors.Wrap(err, "creating complete file in init container")
	}

	logrus.Info("Finished copying build context.")
	return nil
}
