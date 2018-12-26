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

package docker

import (
	"github.com/cedrickring/kbuild/pkg/constants"
	"github.com/cedrickring/kbuild/pkg/util"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
)

//GetConfigAsConfigMap gets the local .docker/config.json in a ConfigMap
func GetConfigAsConfigMap() (*v1.ConfigMap, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.ConfigMapName,
			Labels: map[string]string{
				"builder": "kaniko",
			},
		},
		Data: map[string]string{
			"config.json": string(config),
		},
	}, nil
}

//GetConfig reads the docker config located at ~/.docker/config.json
func GetConfig() ([]byte, error) {
	home := util.HomeDir()
	if home == "" {
		return nil, errors.New("Can't find docker config at ~/.docker/config.json")
	}
	dockerConfigPath := filepath.Join(home, ".docker", "config.json")

	if _, err := os.Stat(dockerConfigPath); os.IsNotExist(err) {
		return nil, errors.New("Can't find docker config at ~/.docker/config.json")
	}

	return ioutil.ReadFile(dockerConfigPath)
}
