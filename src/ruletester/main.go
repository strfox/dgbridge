package main

import (
	"dgbridge/src/lib"
	"encoding/json"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/go-playground/validator/v10"
	"os"
)

var validate *validator.Validate

type CliArgs struct {
	RulesFile string `arg:"required,-r,--rules" help:"Rules to be tested"`
	TestFile  string `arg:"required,-t,--test"  help:"Path to test file"`
}

func main() {
	fmt.Printf("Dgbridge Rule Tester (v%v)\n", lib.Version)

	//
	// Init global state
	//
	validate = validator.New()

	//
	// Parse CLI args
	//
	var args CliArgs
	arg.MustParse(&args)

	//
	// Load files from CLI parameters
	//
	rules, err := loadRulesFile(args)
	if err != nil {
		printError("Failed to load rules file: %v", err)
		os.Exit(1)
	}
	root, err := loadFileRoot(args)
	if err != nil {
		printError("Failed to load test file: %v", err)
		os.Exit(1)
	}

	testRunner := NewTestRunner(root, rules)
	testRunner.RunTests()
}

func loadFileRoot(args CliArgs) (*FileRoot, error) {
	fileContents, err := os.ReadFile(args.TestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load test file: %v", err)
	}
	var test FileRoot
	if err := json.Unmarshal(fileContents, &test); err != nil {
		return nil, fmt.Errorf("error loading test file: %v", err)
	}
	if err := validate.Struct(test); err != nil {
		return nil, fmt.Errorf(
			"Validation of test file failed.\n"+
				"Please look at the errors below and try to fix them.\n"+
				"%v\n", err)
	}
	return &test, nil
}

func loadRulesFile(args CliArgs) (*lib.Rules, error) {
	rules, err := lib.LoadRules(args.RulesFile)
	if err != nil {
		return nil, fmt.Errorf("error loading rules: %v", err)
	}
	if err := validate.Struct(rules); err != nil {
		return nil, fmt.Errorf(
			"Validation of rules file failed.\n"+
				"Please look at the errors below and try to fix them.\n"+
				"%v\n", err)
	}
	return rules, nil
}

func printError(format string, vargs ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, vargs...)
}
