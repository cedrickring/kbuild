package main

import (
	"github.com/cedrickring/kaniko-build/pkg/kaniko"
	"github.com/cedrickring/kaniko-build/pkg/log"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	dockerfile string
	workingDir string
	imageTag   string
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
	rootCmd.Flags().StringVarP(&dockerfile, "dockerfile", "d", ".", "Path to Dockerfile inside working directory")
	rootCmd.Flags().StringVarP(&workingDir, "workDir", "w", ".", "Working directory")
	rootCmd.Flags().StringVarP(&imageTag, "tag", "t", "", "Final image name (required)")
	rootCmd.Flags().BoolVarP(&useCache, "cache", "c", false, "Enable RUN command caching")
	rootCmd.Flags().StringVarP(&cacheRepo, "cache-repo", "", "", "Repository for cached images (see --cache)")
	rootCmd.MarkFlagRequired("tag")

	rootCmd.Execute()
}

func run(_ *cobra.Command, _ []string) {
	_, err := name.NewTag(imageTag, name.WeakValidation) //weak validation to allow only <registry/<repo> without a specific tag
	if err != nil {
		log.Err(err)
		return
	}

	if err := checkForDockerfile(); err != nil {
		log.Err(err)
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
		ImageTag:       imageTag,
		Cache:          useCache,
		CacheRepo:      cacheRepo,
	}
	log.Err(b.StartBuild())
}

func checkForDockerfile() error {
	df := strings.Replace(dockerfile, ".", "Dockerfile", -1)
	if _, err := os.Stat(filepath.Join(workingDir, df)); err != nil {
		return errors.Errorf("Can't find Dockerfile in the working directory. (%s)", workingDir + "/" + df)
	}
	return nil
}
