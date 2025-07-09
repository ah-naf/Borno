package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/ah-naf/borno/token"
)

var HadError bool = false
var HadRuntimeError bool = false

func GlobalError(line int, message string) {
	report(line, "", message)
}

func GlobalErrorToken(t token.Token, message string) {
	if t.Type == token.EOF {
		report(t.Line, " at end", message)
	} else {
		report(t.Line, " at '"+t.Lexeme+"'", message)
	}
}

func report(line int, where, message string) {
	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, message)
	HadError = true
}

func RuntimeError(token token.Token, message string) {
	fmt.Fprintf(os.Stderr, "%s\n[line %d]\n", message, token.Line)
	HadRuntimeError = true
}

func ConvertBanglaDigitsToASCII(input string) string {
	replacements := map[rune]rune{
		'০': '0', '১': '1', '২': '2', '৩': '3', '৪': '4',
		'৫': '5', '৬': '6', '৭': '7', '৮': '8', '৯': '9',
	}

	var result strings.Builder
	for _, r := range input {
		if replacement, exists := replacements[r]; exists {
			result.WriteRune(replacement) // Convert Bangla digit to ASCII
		} else {
			result.WriteRune(r) // Keep other characters unchanged
		}
	}
	return result.String()
}
