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
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/cedrickring/kbuild/pkg/kubernetes"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
)

const credentialsSecretName = "kaniko-gcs-secret"

//GCS represents a google cloud storage build context
type GCS struct {
	Namespace string
	Bucket    string
	Ctx       context.Context

	tar string
}

//Cleanup removes the context from the gcs bucket and removes the gcs secret from the cluster
func (g GCS) Cleanup() {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		logrus.WithError(err).Errorln("error occurred while creating client")
		return
	}

	if err := client.Bucket(g.Bucket).Object(g.tar).Delete(context.Background()); err != nil {
		logrus.WithError(err).Errorln("error occurred while deleting tar from bucket")
	}

	k8sClient, err := kubernetes.GetClient()
	if err != nil {
		logrus.WithError(err).Errorln("error occurred while getting k8s client")
		return
	}

	if err := k8sClient.CoreV1().Secrets(g.Namespace).Delete(credentialsSecretName, &metav1.DeleteOptions{}); err != nil {
		logrus.WithError(err).Errorln("error occurred while deleting gcs secret")
	}
}

//PrepareCredentials creates a v1.Secret with the contents of the Service Account JSON
//found at GOOGLE_APPLICATION_CREDENTIALS
func (g GCS) PrepareCredentials() error {
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		return errors.New("env var GOOGLE_APPLICATION_CREDENTIALS must be set")
	}

	creds, err := ioutil.ReadFile(credsPath)
	if err != nil {
		return errors.Wrap(err, "reading gcs credentials file")
	}

	client, err := kubernetes.GetClient()
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: credentialsSecretName,
			Labels: map[string]string{
				"builder": "kaniko",
			},
		},
		Data: map[string][]byte{
			"kaniko-secret.json": creds,
		},
	}

	if _, err = client.CoreV1().Secrets(g.Namespace).Create(secret); err != nil {
		return errors.Wrap(err, "creating gcs secret")
	}

	return nil
}

//ModifyPod adds the gcs secret as a volume to the pod to access the bucket from Kaniko
func (g GCS) ModifyPod(pod *v1.Pod) {
	//Mount gcs secret as volume
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: "google-credentials",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: credentialsSecretName,
			},
		},
	})

	//Add volume mount for secret
	pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
		Name:      "google-credentials",
		MountPath: "/secret",
	})

	//Add env var to specify path of credentials
	pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, v1.EnvVar{
		Name:  "GOOGLE_APPLICATION_CREDENTIALS",
		Value: "/secret/kaniko-secret.json",
	})
}

//UploadTar uploads the build context to the specified gcs bucket
func (g *GCS) UploadTar(pod *v1.Pod, tarPath string) error {
	client, err := storage.NewClient(g.Ctx)
	if err != nil {
		return errors.Wrap(err, "creating storage client")
	}

	tar, err := os.Open(tarPath)
	if err != nil {
		return errors.Wrap(err, "opening tar")
	}
	defer tar.Close()

	g.tar = filepath.Base(tarPath)
	writer := client.Bucket(g.Bucket).Object(g.tar).NewWriter(g.Ctx)

	if _, err := io.Copy(writer, tar); err != nil {
		return errors.Wrap(err, "copying tar to bucket")
	}

	if err := writer.Close(); err != nil {
		return errors.Wrap(err, "closing bucket writer")
	}

	pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, fmt.Sprintf("--context=gs://%s/%s", g.Bucket, g.tar))
	return nil
}

//RequiresPod always returns false as the pod should not be started before the context is uploaded
func (GCS) RequiresPod() bool {
	return false
}
