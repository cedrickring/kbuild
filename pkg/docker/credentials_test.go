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
	"testing"
)

func TestGuessRegistryFromTag(t *testing.T) {
	tests := []struct {
		imageTag         string
		expectedRegistry string
	}{
		{
			imageTag:         "my.registry.com/tag:latest",
			expectedRegistry: "my.registry.com",
		},
		{
			imageTag:         "registry.com/tag:latest",
			expectedRegistry: "registry.com",
		},
		{
			imageTag:         "openjdk:latest",
			expectedRegistry: "https://index.docker.io/v1/",
		},
	}

	for _, test := range tests {
		registry := GuessRegistryFromTag(test.imageTag)
		if registry != test.expectedRegistry {
			t.Errorf("Expected %s but got %s\n", test.expectedRegistry, registry)
		}
	}
}
