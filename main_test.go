package main_test

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/suite"
)

type MainTestSuite struct {
	suite.Suite
}

func (suite *MainTestSuite) TestMain() {
	testscript.Run(suite.T(), testscript.Params{
		Dir: "testdata",
		Setup: func(e *testscript.Env) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			e.Setenv("HOME", home)
			return exec.Command("go", "build", "-o", e.WorkDir+"/balagan", ".").Run()
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"startsWithUnderscore": startsWithUnderscore,
			"checkSymbol":          checkSymbol,
		},
	})
}

func startsWithUnderscore(ts *testscript.TestScript, neg bool, args []string) {
	filename := args[0]
	line, err := strconv.Atoi(args[1])
	ts.Check(err)

	fileContent := ts.ReadFile(filename)
	ts.Check(err)
	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	for i := 0; i < line; i++ {
		scanner.Scan()
	}
	text := scanner.Text()
	trimmed := strings.Trim(text, " \t")
	if neg && trimmed[0] == '_' {
		ts.Fatalf("%s:%d starts with underscore: %s", filename, line, text)
	}
	if !neg && trimmed[0] != '_' {
		ts.Fatalf("%s:%d does not start with underscore: %s", filename, line, text)
	}
}

func checkSymbol(ts *testscript.TestScript, neg bool, args []string) {
	filename, symbol := args[0], args[1]
	fileContent := ts.ReadFile(filename)
	file, err := parser.ParseFile(token.NewFileSet(), filename, strings.NewReader(fileContent), parser.SkipObjectResolution)
	ts.Check(err)
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if symbol == ident.Name {
				if neg {
					ts.Fatalf("Found symbol %s in %s:\n\n%s", symbol, filename, fileContent)
				}
				found = true
				return false
			}
		}
		return true
	})
	if !found && !neg {
		ts.Fatalf("Symbol %s not found in %s:\n\n%s", symbol, filename, fileContent)
	}
}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
