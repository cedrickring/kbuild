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

	"github.com/cedrickring/kbuild/pkg/constants"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b Build) getKanikoPod() *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "kaniko-",
			Labels: map[string]string{
				"builder": "kaniko",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  constants.KanikoContainerName,
					Image: "gcr.io/kaniko-project/executor",
					Args: []string{
						"--dockerfile=" + b.DockerfilePath,
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "docker-config",
							MountPath: "/kaniko/.docker",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				{
					Name: "docker-config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: constants.ConfigMapName,
							},
						},
					},
				},
			},
		},
	}

	//add all image tags as destination arguments
	for _, tag := range b.ImageTags {
		pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, fmt.Sprintf("--destination=%s", tag))
	}

	//add all build args to Kaniko container args
	for _, arg := range b.BuildArgs {
		pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, fmt.Sprintf("--build-arg=%s", arg))
	}

	//only enable caching if provided by "kbuild --cache"
	if b.Cache {
		cacheRepo := b.CacheRepo
		if cacheRepo == "" {
			tag, err := name.NewTag(b.ImageTags[0], name.WeakValidation)
			if err != nil {
				return nil
			}

			cacheRepo = fmt.Sprintf("%s/%scache", tag.RegistryStr(), tag.RepositoryStr())
		}

		pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, "--cache=true", fmt.Sprintf("--cache-repo=%s", cacheRepo))
	}

	return pod
}
