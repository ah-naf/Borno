package utils

import (
	"fmt"

	"github.com/ah-naf/crafting-interpreter/token"
)

var HadError bool = false

func GlobalError(line int, message string) {
	report(line, "", message)
}

func GlobalErrorToken(t token.Token, message string) {
	if t.Type == token.EOF {
		report(t.Line, " at end", message)
	} else {
		report(t.Line, " at '" + t.Lexeme + "'", message)
	}
}

func report(line int, where, message string) {
	fmt.Printf("[line %d] Error %s: %s\n", line, where, message)
	HadError = true
}
