package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ah-naf/borno/interpreter"
	"github.com/ah-naf/borno/lexer"
	"github.com/ah-naf/borno/parser"
	"github.com/ah-naf/borno/utils"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: borno [script]")
		os.Exit(64)
	} else if len(os.Args) == 2 {
		scriptFile := os.Args[1]

		// Extract the file extension.
		ext := filepath.Ext(scriptFile) // e.g. ".bn" or ".borno"

		if ext != ".bn" {
			fmt.Println("Invalid file extension. Please use `.bn` for Borno scripts.")
			os.Exit(64)
		}

		runFile(scriptFile)
	} else {
		runPrompt()
	}
}

func runFile(path string) {
	rawContent, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	run(string(rawContent), false)

	if utils.HadError {
		os.Exit(65)
	}
	if utils.HadRuntimeError {
		os.Exit(70)
	}
}

func runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf(">> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		run(line, true)

		utils.HadError = false
		utils.HadRuntimeError = false
	}
}

func run(source string, isRepl bool) {
	runeSource := []rune(source)
	scanner := lexer.NewScanner(runeSource)
	tokens := scanner.ScanTokens()
	// fmt.Printf("%#v\n", tokens)

	Parser := parser.NewParser(tokens)
	expr, _ := Parser.Parse()

	if utils.HadError {
		return
	}

	interpreter := interpreter.NewInterpreter()
	interpreter.Interpret(expr, isRepl)
	if utils.HadRuntimeError {
		return
	}

	// for _, stmt := range expr {
	// 	// prettyPrint(stmt) // Use %#v to print all the nested fields and structs
	// 	fmt.Println(stmt)

	// }
	// fmt.Println(expr)
}

func prettyPrint(v interface{}) {
	// Marshal the struct to JSON with indentation
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Print the resulting JSON string
	fmt.Println(string(data))
}
