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
	"github.com/cedrickring/kbuild/pkg/kaniko"
	"github.com/cedrickring/kbuild/pkg/log"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	dockerfile string
	workingDir string
	imageTags  []string
	useCache   bool
	cacheRepo  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "kbuild",
		Example: "kbuild -t <repo>:<tag>",
		Short:   "Build a container image inside a Kubernetes Cluster with Kaniko.",
		Run:     run,
	}
	rootCmd.Flags().StringVarP(&dockerfile, "dockerfile", "d", "Dockerfile", "Path to Dockerfile inside working directory")
	rootCmd.Flags().StringVarP(&workingDir, "workDir", "w", ".", "Working directory")
	rootCmd.Flags().StringSliceVarP(&imageTags, "tag", "t", nil, "Final image tag(s) (required)")
	rootCmd.Flags().BoolVarP(&useCache, "cache", "c", false, "Enable RUN command caching")
	rootCmd.Flags().StringVarP(&cacheRepo, "cache-repo", "", "", "Repository for cached images (see --cache)")
	rootCmd.MarkFlagRequired("tag")

	rootCmd.Execute()
}

func run(_ *cobra.Command, _ []string) {
	if err := validateImageTags(); err != nil {
		log.Err(err)
		os.Exit(1)
		return
	}

	if err := checkForDockerfile(); err != nil {
		log.Err(err)
		os.Exit(1)
		return
	}

	cachingInfo := "Run-Step caching is %s."
	if useCache {
		log.Infof(cachingInfo, "enabled")
	} else {
		log.Infof(cachingInfo, "disabled")
	}

	b := kaniko.Build{
		DockerfilePath: dockerfile,
		WorkDir:        workingDir,
		ImageTags:      imageTags,
		Cache:          useCache,
		CacheRepo:      cacheRepo,
	}
	err := b.StartBuild()
	if err != nil {
		if err == kaniko.ErrorBuildFailed {
			log.Error("Build failed.")
		} else {
			log.Err(err)
		}
		os.Exit(1)
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
