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

package util

import (
	"crypto/rand"
	"fmt"
	"os"
)

//HomeDir returns the %USERPROFILE% directory on windows or $HOME directory on Unix
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" { //unix
		return h
	}
	return os.Getenv("USERPROFILE") //windows
}

//RandomID creates a 32 character random ID
func RandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
