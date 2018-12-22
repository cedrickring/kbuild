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
		buildArgs     []string
		shouldErr     bool
	}{
		{
			dockerfile:    "Dockerfile.test",
			expectedPaths: []string{"test/test.go", "test/Dockerfile.test"},
			shouldErr:     false,
		},
		{
			dockerfile:    "Dockerfile.dot-test",
			expectedPaths: []string{"test", "test/Dockerfile.dot-test"},
			shouldErr:     false,
		},
		{
			dockerfile:    "Dockerfile.arg-test",
			expectedPaths: []string{"test/test.go", "test/Dockerfile.arg-test"},
			buildArgs:     []string{"TEST=test.go"},
			shouldErr:     false,
		},
		{
			dockerfile: "Dockerfile.arg-test",
			shouldErr:  true,
		},
		{
			dockerfile:    "Dockerfile.arg-env-test",
			expectedPaths: []string{"test/test.go", "test/Dockerfile.arg-env-test"},
		},
	}

	for _, test := range tests {
		paths, err := GetFilePaths("test", test.dockerfile, test.buildArgs)
		if err != nil {
			if !test.shouldErr {
				t.Errorf("Couldn't get file paths: %s", err.Error())
			}
			continue
		}

		if !reflect.DeepEqual(test.expectedPaths, paths) {
			t.Errorf("Expected %s but got %s", strings.Join(test.expectedPaths, ","), strings.Join(paths, ","))
		}
	}
}

func TestGetBuildArgs(t *testing.T) {
	var tests = []struct {
		buildArgs []string
		argsMap   map[string]string
		shouldErr bool
	}{
		{
			buildArgs: []string{"ARG1=VALUE", "ARG2=VALUE2"},
			argsMap:   map[string]string{"ARG1": "VALUE", "ARG2": "VALUE2"},
			shouldErr: false,
		},
		{
			buildArgs: []string{"no-key-value"},
			argsMap:   make(map[string]string),
			shouldErr: true,
		},
	}

	for _, test := range tests {
		args, err := parseBuildArgs(test.buildArgs)
		if err != nil {
			if !test.shouldErr {
				t.Errorf("Expected parseBuildArgs not to error but got error %s", err)
			}
			continue
		}

		if args == nil || !reflect.DeepEqual(args, test.argsMap) {
			t.Errorf("Expected %s but got %s", test.argsMap, args)
		}
	}

}
