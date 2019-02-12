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

//CreateContextFromWorkingDir creates a build context of the provided directory and writes it to the Writer
func CreateContextFromWorkingDir(workDir, dockerfile string, w io.Writer, buildArgs []string) error {
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return errors.Wrap(err, "get context from workDir")
	}

	paths, err := GetFilePaths(workDir, dockerfile, buildArgs) //paths are relative to the directory this executable runs in
	if err != nil {
		return err
	}

	pm, err := excludeMatcher(workDir)
	if err != nil {
		return errors.Wrap(err, "getting exclude pattern matcher")
	}

	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for _, path := range paths {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return err
		}

		if info.IsDir() {
			err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if p == path {
					return nil
				}

				rel, err := filepath.Rel(workDir, p) //make path relative to work dir
				if err != nil {
					return err
				}

				if skip, err := shouldSkipPath(pm, info.IsDir(), rel); skip {
					return err
				}

				if !info.Mode().IsRegular() {
					return nil
				}

				header, err := tar.FileInfoHeader(info, info.Name())
				if err != nil {
					return errors.Wrap(err, "creating tar file info header")
				}
				header.Name = filepath.ToSlash(rel)

				return copyFile(header, p, tw)
			})
			if err != nil {
				return err
			}
		} else {
			if skip, _ := shouldSkipPath(pm, info.IsDir(), path); skip {
				continue
			}

			if !info.Mode().IsRegular() {
				continue
			}

			rel, err := filepath.Rel(workDir, path) //make path relative to work dir
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return errors.Wrap(err, "creating tar file info header")
			}
			header.Name = filepath.ToSlash(rel)

			if err := copyFile(header, path, tw); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(header *tar.Header, path string, to *tar.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "opening file")
	}

	err = to.WriteHeader(header)
	if err != nil {
		return errors.Wrap(err, "writing tar file header")
	}

	if _, err := io.Copy(to, f); err != nil {
		return errors.Wrap(err, "copying file into tar")
	}

	return nil
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
