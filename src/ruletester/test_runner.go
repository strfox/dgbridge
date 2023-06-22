package main

import (
	"dgbridge/src/lib"
	"fmt"
	"strings"
)

type TestRunner struct {
	TestFile *FileRoot
	Rules    *lib.Rules
}

type TestResults struct {
	Passed int
	Failed int
}

type Test interface {
	Run(testRunner *TestRunner, number int, rules *lib.Rules) bool
}

func NewTestRunner(testFile *FileRoot, rules *lib.Rules) TestRunner {
	return TestRunner{
		TestFile: testFile,
		Rules:    rules,
	}
}

func (r *TestRunner) RunTests() {
	results := TestResults{
		Passed: 0,
		Failed: 0,
	}
	results.Add(RunTests(r, "SubprocessToDiscord", r.TestFile.Tests.SubprocessToDiscord, r.Rules))
	results.Add(RunTests(r, "DiscordToSubprocess", r.TestFile.Tests.DiscordToSubprocess, r.Rules))

	fmt.Printf("Finished: Tests passed: %v, failed: %v\n", results.Passed, results.Failed)
}

func RunTests[T Test](testRunner *TestRunner, bannerTitle string, tests []T, rules *lib.Rules) TestResults {
	results := TestResults{
		Passed: 0,
		Failed: 0,
	}

	printBanner(bannerTitle, len(tests))

	for i, test := range tests {
		pass := test.Run(testRunner, i, rules)
		if pass {
			results.Passed++
		} else {
			results.Failed++
		}
	}
	return results
}

func printBanner(bannerTitle string, amountTests int) {
	banner := fmt.Sprintf("%v tests: Running %v tests\n", bannerTitle, amountTests)
	{
		line := strings.Repeat("-", len(banner))
		banner = line + "\n" + banner + line + "\n"
	}
	fmt.Printf(banner)
}

func (t SubprocessToDiscordTest) Run(_ *TestRunner, number int, rules *lib.Rules) bool {
	result := lib.ApplyRules(rules.SubprocessToDiscord, nil, t.Input)
	if result != t.Expect {
		fmt.Printf(
			"❌  SubprocessToDiscordTest Test #%v: FAIL:\n"+
				"\tInput:\t\t%v\n"+
				"\tExpected:\t%v\n"+
				"\tGot:\t\t%v\n",
			number, t.Input, t.Expect, result,
		)
		return false
	}
	fmt.Printf("✅  Test #%v: PASS\n", number)
	return true
}

func (t DiscordToSubprocessTest) Run(testRunner *TestRunner, number int, rules *lib.Rules) bool {
	userProps, ok := testRunner.TestFile.UserProps[t.UserProps]
	if !ok {
		printError("❌  Test #%v: bad test: missing UserProps \"%v\".\n", number, t.UserProps)
		return false
	}

	result := lib.ApplyRules(rules.DiscordToSubprocess, &userProps, t.Input)
	if result != t.Expect {
		fmt.Printf(
			"❌  d2s Test #%v: FAIL:\n"+
				"\tInput:\t\t%v\n"+
				"\tExpected:\t%v\n"+
				"\tGot:\t\t%v\n",
			number, t.Input, t.Expect, result,
		)
		return false
	}
	fmt.Printf("✅  Test #%v: PASS\n", number)
	return true
}

func (r *TestResults) Add(other TestResults) {
	r.Passed += other.Passed
	r.Failed += other.Failed
}
