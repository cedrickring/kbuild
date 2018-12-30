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

package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cedrickring/kbuild/pkg/docker"
	"github.com/cedrickring/kbuild/pkg/kaniko"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"os"
	"path/filepath"
)

var (
	dockerfile string
	workingDir string
	cacheRepo  string
	namespace  string
	imageTags  []string
	buildArgs  []string
	useCache   bool
	username   string
	password   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "kbuild",
		Example: "kbuild -t <repo>:<tag>",
		Short:   "Build a container image inside a Kubernetes Cluster with Kaniko.",
		Run:     run,
	}
	rootCmd.Flags().StringVarP(&dockerfile, "dockerfile", "d", "Dockerfile", "Path to Dockerfile inside working directory")
	rootCmd.Flags().StringVarP(&workingDir, "workdir", "w", ".", "Working directory")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The namespace to run the build in")
	rootCmd.Flags().StringVarP(&cacheRepo, "cache-repo", "", "", "Repository for cached images (see --cache)")
	rootCmd.Flags().StringVarP(&username, "username", "u", "", "Docker Registry username")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "Docker Registry password")
	rootCmd.Flags().StringSliceVarP(&imageTags, "tag", "t", nil, "Final image tag(s) (required)")
	rootCmd.Flags().StringSliceVarP(&buildArgs, "build-arg", "", nil, "Optional build arguments (ARG)")
	rootCmd.Flags().BoolVarP(&useCache, "cache", "c", false, "Enable RUN command caching")
	rootCmd.MarkFlagRequired("tag")

	rootCmd.Execute()
}

func run(_ *cobra.Command, _ []string) {
	setupLogrus()

	if err := validateImageTags(); err != nil {
		logrus.Fatal(err)
		return
	}

	if err := checkForDockerfile(); err != nil {
		logrus.Fatal(err)
		return
	}

	credentialsMap, err := getCredentialsConfigMap()
	if err != nil {
		logrus.Fatal(err)
		return
	}

	cachingInfo := "Run-Step caching is %s."
	if useCache {
		logrus.Infof(cachingInfo, "enabled")
	} else {
		logrus.Infof(cachingInfo, "disabled")
	}

	logrus.Infof(`Running in namespace "%s"`, namespace)

	b := kaniko.Build{
		DockerfilePath: dockerfile,
		WorkDir:        workingDir,
		ImageTags:      imageTags,
		Cache:          useCache,
		CacheRepo:      cacheRepo,
		Namespace:      namespace,
		BuildArgs:      buildArgs,
		CredentialsMap: credentialsMap,
	}
	err = b.StartBuild()
	if err != nil {
		if err == kaniko.ErrorBuildFailed {
			logrus.Fatal("Build failed.")
		} else {
			logrus.Fatal(err)
		}
	}
}

func validateImageTags() error {
	for _, tag := range imageTags {
		_, err := name.NewTag(tag, name.WeakValidation) //weak validation to allow only <registry/<repo> without a specific tag
		if err != nil {
			return err
		}
	}
	return nil
}

func checkForDockerfile() error {
	if _, err := os.Stat(filepath.Join(workingDir, dockerfile)); err != nil {
		return errors.Errorf("Can't find Dockerfile in the working directory. (%s/%s)", workingDir, dockerfile)
	}
	return nil
}

func setupLogrus() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logrus.SetOutput(os.Stdout)
}

func getCredentialsConfigMap() (*v1.ConfigMap, error) {
	var credentials []byte

	//check if credentials have been provided by flags
	if username != "" && password != "" {
		logrus.Infoln("Using credentials from flags")
		registry := docker.GuessRegistryFromTag(imageTags[0])
		creds, err := docker.GetCredentialsFromFlags(username, password, registry)
		if err != nil {
			return nil, errors.Wrap(err, "getting credentials from flags")
		}
		credentials = creds
	} else { //otherwise read ~/.docker/config.json
		logrus.Infoln("Using credentials from ~/.docker/config.json")
		creds, err := docker.GetCredentialsFromConfig()
		if err != nil {
			return nil, errors.Wrap(err, "getting credentials from config")
		}
		credentials = creds
	}

	return docker.GetCredentialsAsConfigMap(credentials), nil
}
