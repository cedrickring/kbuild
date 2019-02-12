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
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//GetFilePaths returns all paths required to build the docker image
func GetFilePaths(workDir, dockerfile string, buildArgs []string) ([]string, error) {
	dfPath := filepath.ToSlash(filepath.Join(workDir, dockerfile))
	f, err := os.Open(dfPath)
	if err != nil {
		return nil, errors.Wrap(err, "opening dockerfile")
	}

	result, err := parser.Parse(f)
	if err != nil {
		return nil, errors.Wrap(err, "parsing dockerfile")
	}

	var paths []string

	children := result.AST.Children

	envVars := make(map[string]string)
	args, err := parseBuildArgs(buildArgs)
	if err != nil {
		return nil, errors.Wrap(err, "parsing build args from flags")
	}

	for _, node := range children {
		switch node.Value {
		case command.Copy, command.Add:
			parsed, err := parseCopyOrAdd(workDir, node, envVars, args)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("Dockerfile line %d", node.StartLine))
			}
			paths = append(paths, parsed...)
		case command.Env:
			envVars[node.Next.Value] = node.Next.Next.Value
		case command.Arg:
			err := parseArgCommand(node, args)
			if err != nil {
				return nil, err
			}
		}
	}

	paths = append(paths, dfPath) //add Dockerfile every time

	return paths, nil
}

func parseCopyOrAdd(wd string, node *parser.Node, envVars map[string]string, buildArgs map[string]string) ([]string, error) {
	var paths []string

	for _, flag := range node.Flags {
		if strings.HasPrefix(flag, "--from") { //ignore paths copied from another build step
			return paths, nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "getting workdir")
	}

	lex := shell.NewLex(rune('\\'))
	for node = node.Next; node.Next != nil; node = node.Next {
		if match, err := regexp.MatchString("^https?://(.*)", node.Value); err != nil || match {
			logrus.Infof("Skipping external dependency %s", node.Value)
			continue //skip external dependencies
		}

		if !filepath.IsAbs(wd) {
			wd, err = filepath.Abs(wd)
			if err != nil {
				return nil, err
			}
		}
		abs := filepath.ToSlash(filepath.Join(wd, node.Value)) //need forward-slashes in windows so they don't get escaped by lex

		for key, value := range buildArgs {
			if _, ok := envVars[key]; ok {
				continue //ENV variables always override ARG variables
			}

			r := regexp.MustCompile(`(\$` + key + `|\${` + key + `})`)
			submatch := r.FindStringSubmatch(abs)
			if len(submatch) > 0 {
				abs = r.ReplaceAllString(abs, value)
			}
		}

		expanded, err := lex.ProcessWordWithMap(abs, envVars)
		if err != nil {
			return nil, err
		}

		matches, err := filepath.Glob(expanded)
		if err != nil {
			return nil, err
		}

		var relPaths []string
		for _, match := range matches {
			rel, err := filepath.Rel(wd, match) //make path relative to work dir to check for paths outside the build context
			if err != nil || strings.HasPrefix(rel, "..") {
				return nil, errors.Errorf("path %s is not inside the build context", node.Value)
			}

			rel, err = filepath.Rel(cwd, match) //make path relative to cwd so "." gets interpreted correctly
			if err != nil {
				return nil, err
			}
			rel = filepath.ToSlash(rel)

			relPaths = append(relPaths, rel)
		}

		paths = append(paths, relPaths...)
	}

	return paths, nil
}

func parseArgCommand(node *parser.Node, args map[string]string) error {
	arg := node.Next.Value
	if _, ok := args[arg]; ok {
		return nil //skip if arg is set by flag, otherwise check for default value
	}

	if !strings.Contains(arg, "=") {
		if _, ok := args[arg]; !ok {
			return errors.Errorf("required arg %s was not set by --arg flag", arg)
		}
		return nil
	}

	key, value, err := parseArg(arg)
	if err != nil {
		return errors.Wrap(err, "parsing ARG command")
	}
	args[key] = value
	return nil
}

func parseBuildArgs(buildArgs []string) (map[string]string, error) {
	args := make(map[string]string)

	for _, arg := range buildArgs {
		key, value, err := parseArg(arg)
		if err != nil {
			return args, err
		}
		args[key] = value
	}

	return args, nil
}

func parseArg(arg string) (key, value string, err error) {
	split := strings.Split(arg, "=")
	if len(split) != 2 {
		return "", "", errors.New("invalid arg format, must be ARG=VALUE")
	}
	key = split[0]
	value = split[1]
	return key, value, nil
}
