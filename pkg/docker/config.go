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
