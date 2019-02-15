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
	"path/filepath"

	"github.com/cedrickring/kbuild/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" //auth for GKE clusters
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//GetClient creates a new kubernetes client with the kubeconfig at ~/.kube/config
func GetClient() (*kubernetes.Clientset, error) {
	config, err := GetRestConfig()
	if err != nil {
		return nil, errors.Wrap(err, "getting rest config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "new clientset from config")
	}

	return clientset, nil
}

//GetRestConfig returns a new rest.Config based on the kubeconfig at ~/.kube/config
func GetRestConfig() (*rest.Config, error) {
	home := util.HomeDir()
	if home == "" {
		return nil, errors.New("Can't find kubeconfig at ~/.kube/config")
	}
	kubeconfigPath := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "build client config from flags")
	}
	return config, nil
}
