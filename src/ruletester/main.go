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

type (
	FileRoot struct {
		Tests     RuleTest             `validate:"required"`
		UserProps map[string]lib.Props `validate:"dive"`
	}
	RuleTest struct {
		Dis2Srv Dis2SrvTestBlock  `validate:"required"`
		Srv2Dis []Srv2DisTestCase `validate:"required,dive"`
	}
	Dis2SrvTestBlock struct {
		Tests []Dis2SrvTestCase `validate:"required,dive"`
	}
	Dis2SrvTestCase struct {
		Input     string `validate:"required"`
		Expect    string `validate:"required"`
		UserProps string `validate:"required"`
	}
	Srv2DisTestCase struct {
		Input  string `validate:"required"`
		Expect string `validate:"required"`
	}
)

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
	tests := root.Tests

	//
	// Run test group
	//
	passedTests := 0
	failedTests := 0

	fmt.Printf(
		"-------------------------------------------\n"+
			"Srv2Dis tests: Running %v tests\n"+
			"-------------------------------------------\n",
		len(tests.Srv2Dis),
	)
	for i, test := range tests.Srv2Dis {
		result := lib.ApplyRules(rules.SubprocessToDiscord, nil, test.Input)
		if result != test.Expect {
			fmt.Printf(
				"❌  Srv2DisTestCase Test #%v: FAIL:\n"+
					"\tInput:\t\t%v\n"+
					"\tExpected:\t%v\n"+
					"\tGot:\t\t%v\n",
				i, test.Input, test.Expect, result,
			)
			failedTests += 1
			continue
		}
		fmt.Printf("✅  Test #%v: PASS\n", i)
		passedTests += 1
	}

	fmt.Printf(
		"-------------------------------------------\n"+
			"Dis2Srv tests: Running %v tests\n"+
			"-------------------------------------------\n",
		len(tests.Dis2Srv.Tests),
	)
	for i, test := range tests.Dis2Srv.Tests {
		userProps, ok := root.UserProps[test.UserProps]
		if !ok {
			printError("❌  Test #%v: bad test: missing UserProps \"%v\".\n", i, test.UserProps)
			failedTests += 1
			continue
		}

		result := lib.ApplyRules(rules.DiscordToSubprocess, &userProps, test.Input)
		if result != test.Expect {
			fmt.Printf(
				"❌  d2s Test #%v: FAIL:\n"+
					"\tInput:\t\t%v\n"+
					"\tExpected:\t%v\n"+
					"\tGot:\t\t%v\n",
				i, test.Input, test.Expect, result,
			)
			failedTests += 1
			continue
		}
		passedTests += 1
		fmt.Printf("✅  Test #%v: PASS\n", i)
	}

	fmt.Printf("Finished: Tests passed: %v, failed: %v\n", passedTests, failedTests)
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
