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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cedrickring/kbuild/pkg/constants"
	"github.com/cedrickring/kbuild/pkg/util"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dockerConfig struct {
	Auths map[string]auth `json:"auths"`
}

type auth struct {
	Auth string `json:"auth"`
}

var registryRegex = regexp.MustCompile(`^((.*)\.)?(.*)\.(.*)/`)

//GuessRegistryFromTag guesses the container registry based on the provided image tag.
//Defaults to the index.docker.io registry
func GuessRegistryFromTag(imageTag string) string {
	if registryRegex.MatchString(imageTag) {
		return imageTag[:strings.Index(imageTag, "/")]
	}

	return "https://index.docker.io/v1/"
}

//GetCredentialsFromFlags creates a dockerconfig "auths" object containing the provided credentials
func GetCredentialsFromFlags(username, password, registry string) ([]byte, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	config := dockerConfig{
		Auths: map[string]auth{
			registry: {Auth: encoded},
		},
	}

	return json.Marshal(config)
}

//GetCredentialsFromConfig reads the docker config located at ~/.docker/config.json
func GetCredentialsFromConfig() ([]byte, error) {
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

//GetCredentialsAsConfigMap creates a new v1.ConfigMap with the provided credentials
func GetCredentialsAsConfigMap(credentials []byte) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.ConfigMapName,
			Labels: map[string]string{
				"builder": "kaniko",
			},
		},
		Data: map[string]string{
			"config.json": string(credentials),
		},
	}
}
