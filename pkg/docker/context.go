package docker

import (
	"archive/tar"
	"compress/gzip"
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

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return errors.Wrap(err, "creating tar file header")
		}

		//make file path relative inside tar
		if dir == "." {
			header.Name = path
		} else {
			header.Name = strings.Replace(path, dir, "", -1)
			if header.Name[0] == '\\' || header.Name[0] == '/' { //remove leading slashes
				header.Name = header.Name[1:]
			}
		}
		header.Name = strings.Replace(header.Name, "\\", "/", -1) //replace all backslashes with forward slashes

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
