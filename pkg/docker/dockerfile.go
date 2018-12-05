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
	"github.com/cedrickring/kbuild/pkg/log"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func GetFilePaths(workDir, dockerfile string) ([]string, error) {
	f, err := os.Open(filepath.Join(workDir, dockerfile))
	if err != nil {
		return nil, errors.Wrap(err, "opening dockerfile")
	}

	result, err := parser.Parse(f)
	if err != nil {
		return nil, errors.Wrap(err, "parsing dockerfile")
	}

	var paths []string

	children := result.AST.Children

	envVars := map[string]string{}
	for _, node := range children {
		switch node.Value {
		case command.Copy, command.Add:
			paths = append(paths, parseCopyOrAdd(workDir, node, envVars)...)
		case command.Env:
			envVars[node.Next.Value] = node.Next.Next.Value
		}
	}

	fmt.Println(paths)

	return paths, nil
}

func parseCopyOrAdd(wd string, node *parser.Node, envVars map[string]string) []string {
	var paths []string

	for _, flag := range node.Flags {
		if strings.HasPrefix(flag, "--from") { //ignore paths copied from another build step
			return paths
		}
	}

	lex := shell.NewLex(rune('\\'))

Node:
	for node = node.Next; node.Next != nil; node = node.Next {
		abs := strings.Replace(filepath.Join(wd, node.Value), "\\", "/", -1) //need forward-slashes in windows so they don't get escaped by lex

		expanded, err := lex.ProcessWordWithMap(abs, envVars)
		if err != nil {
			log.Err(err)
			continue
		}

		matches, err := filepath.Glob(expanded)
		if err != nil {
			log.Err(err)
			continue
		}

		var relPaths []string
		for _, match := range matches {
			rel, err := filepath.Rel(wd, match)
			if err != nil || strings.HasPrefix(rel, "..") {
				log.Err(err)
				continue Node
			}

			relPaths = append(relPaths, rel)
		}

		paths = append(paths, relPaths...)
	}

	return paths
}
