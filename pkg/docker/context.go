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
	"archive/tar"
	"compress/gzip"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//GetContextFromDir creates a build context of the provided directory and writes it to the Writer
func GetContextFromDir(dir string, w io.Writer) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return errors.Wrap(err, "get context from dir")
	}

	pm, err := excludeMatcher(dir)
	if err != nil {
		return errors.Wrap(err, "getting exclude pattern matcher")
	}

	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dir {
			return nil //do not include the original path in the tar.gz
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}

		if ret, err := shouldSkipPath(pm, info.IsDir(), relPath); ret {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return errors.Wrap(err, "creating tar file header")
		}

		header.Name = strings.Replace(relPath, "\\", "/", -1) //replace all backslashes with forward slashes

		if err := tw.WriteHeader(header); err != nil {
			return errors.Wrap(err, "write header to tar")
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "opening file")
		}

		if _, err := io.Copy(tw, f); err != nil {
			return errors.Wrap(err, "copying file into tar")
		}

		f.Close()

		return nil
	})
}

func shouldSkipPath(pm *fileutils.PatternMatcher, isDir bool, relPath string) (bool, error) {
	skip, err := pm.Matches(relPath)
	if err != nil {
		return true, errors.Wrap(err, "matching exclude pattern")
	}

	if skip {
		if !isDir {
			return true, nil
		}

		if !pm.Exclusions() {
			return true, filepath.SkipDir
		}

		dirSlash := relPath + string(filepath.Separator)

		for _, pat := range pm.Patterns() {
			if !pat.Exclusion() {
				continue
			}
			if strings.HasPrefix(pat.String()+string(filepath.Separator), dirSlash) {
				// found a match - so can't skip this dir
				return true, nil
			}
		}

		// No matching exclusion dir so just skip dir
		return true, filepath.SkipDir
	}

	return false, nil
}

func excludeMatcher(dir string) (*fileutils.PatternMatcher, error) {
	var excludes []string

	f, err := os.Open(filepath.Join(dir, ".dockerignore"))
	switch {
	case os.IsNotExist(err):
		return fileutils.NewPatternMatcher(excludes)
	case err != nil:
		return nil, err
	}
	defer f.Close()

	excludes, err = dockerignore.ReadAll(f)
	if err != nil {
		return nil, err
	}

	excludes = append(excludes, ".dockerignore")

	return fileutils.NewPatternMatcher(excludes)
}
