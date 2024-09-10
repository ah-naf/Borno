package utils

import "fmt"

var HadError bool = false

func GlobalError(line int, message string) {
	report(line, "", message)
}

func report(line int, where, message string) {
	fmt.Printf("[line %d] Error %s: %s\n", line, where, message)
	HadError = true
}
