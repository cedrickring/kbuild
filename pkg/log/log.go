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

//Err prints an error it's not nil (never panics)
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
