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
	"github.com/cedrickring/kbuild/pkg/utils"
	"github.com/cedrickring/kbuild/pkg/utils/constants"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GetConfigAsConfigMap gets the local .docker/config.json in a ConfigMap
func GetConfigAsConfigMap() (*v1.ConfigMap, error) {
	config, err := utils.GetDockerConfig()
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
