package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

//GetDockerConfig reads the docker config located at ~/.docker/config.json
func GetDockerConfig() ([]byte, error) {
	home := homeDir()
	if home == "" {
		return nil, errors.New("Can't find docker config at ~/.docker/config.json")
	}
	dockerConfigPath := filepath.Join(home, ".docker", "config.json")

	if _, err := os.Stat(dockerConfigPath); os.IsNotExist(err) {
		return nil, errors.New("Can't find docker config at ~/.docker/config.json")
	}

	return ioutil.ReadFile(dockerConfigPath)
}

//RandomID creates a 32 character random ID
func RandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
