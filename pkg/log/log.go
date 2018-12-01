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

package log

import (
	"fmt"
	a "github.com/logrusorgru/aurora"
)

//Infof prints a formatted info message
func Infof(message string, params ...interface{}) {
	if len(params) > 0 {
		fmt.Print(a.Green("[INFO] "))
		fmt.Printf(message+"\n", params...)
	}
}

//Info prints an info message similar to fmt.Println(i...)
func Info(i ...interface{}) {
	fmt.Print(a.Green("[INFO] "))
	fmt.Println(i...)
}

//Errorf prints a formatted error message
func Errorf(message string, params ...interface{}) {
	if len(params) > 0 {
		fmt.Print(a.Red("[ERROR] "))
		fmt.Printf(message+"\n", params...)
	}
}

//Error prints an error message similar to fmt.Println(i...)
func Error(i ...interface{}) {
	fmt.Print(a.Red("[ERROR] "))
	fmt.Println(i...)
}

//Err prints an error if it's not nil (never panics)
func Err(err error) {
	if err != nil {
		fmt.Print(a.Red("[ERROR] "))
		fmt.Println(err.Error())
	}
}

//Warnf prints a formatted warning message
func Warnf(message string, params ...interface{}) {
	if len(params) > 0 {
		fmt.Print(a.Brown("[WARN] "))
		fmt.Printf(message+"\n", params...)
	}
}

//Warn prints a warning message similar to fmt.Println(i...)
func Warn(i ...interface{}) {
	fmt.Print(a.Brown("[WARN] "))
	fmt.Println(i...)
}
