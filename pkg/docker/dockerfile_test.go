package docker

import (
	"reflect"
	"strings"
	"testing"
)

func TestGetFilePaths(t *testing.T) {
	var tests = []struct {
		dockerfile    string
		expectedPaths []string
	}{
		{
			dockerfile:    "Dockerfile.test",
			expectedPaths: []string{"test/test.go", "test/Dockerfile.test"},
		},
		{
			dockerfile:    "Dockerfile.dot-test",
			expectedPaths: []string{"test", "test/Dockerfile.dot-test"},
		},
	}

	for _, test := range tests {
		paths, err := GetFilePaths("test", test.dockerfile)
		if err != nil {
			t.Errorf("Couldn't get file paths: %s", err.Error())
			continue
		}

		if !reflect.DeepEqual(test.expectedPaths, paths) {
			t.Errorf("Expected %s but got %s", strings.Join(test.expectedPaths, ","), strings.Join(paths, ","))
		}
	}
}
