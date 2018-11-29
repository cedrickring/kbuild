package kaniko

import (
	"fmt"
	"github.com/cedrickring/kaniko-build/pkg/utils/constants"
	"github.com/google/go-containerregistry/pkg/name"
	"k8s.io/api/core/v1"
	"strings"
)
import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func (b Build) getKanikoPod() *v1.Pod {
	dockerfile := strings.Replace(b.DockerfilePath, ".", "Dockerfile", -1)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "kaniko-",
			Labels: map[string]string{
				"builder": "kaniko",
			},
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
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
							MountPath: "/kaniko/build-context",
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:  "kaniko-build",
					Image: "gcr.io/kaniko-project/executor",
					Args: []string{
						"--dockerfile=" + dockerfile,
						"--context=dir:///kaniko/build-context",
						"--destination=" + b.ImageTag,
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "docker-config",
							MountPath: "/kaniko/.docker",
						},
						{
							Name:      "build-context",
							MountPath: "/kaniko/build-context",
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
				{
					Name: "build-context",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	//only enable caching if provided by "kbuild --cache"
	if b.Cache {
		cacheTag := b.CacheRepo
		if cacheTag == "" {
			tag, err := name.NewTag(b.ImageTag, name.WeakValidation)
			if err != nil {
				return nil
			}

			cacheTag = fmt.Sprintf("%s/%scache", tag.RegistryStr(), tag.RepositoryStr())
		}

		pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, "--cache=true", "--cache-repo="+cacheTag)
	}

	return pod
}
