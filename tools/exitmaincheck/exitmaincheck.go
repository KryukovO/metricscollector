package exitmaincheck

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitmaincheck",
	Doc:  "checks os.Exit calls in main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	isExit := func(call *ast.CallExpr) bool {
		if fn, ok := call.Fun.(*ast.SelectorExpr); ok {
			p, ok1 := fn.X.(*ast.Ident)

			if !ok1 || p.Name != "os" || fn.Sel.Name != "Exit" {
				return false
			}

			return true
		}

		return false
	}

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename

		// Игнорирование генерируемых файлов тестов
		if !strings.HasSuffix(filename, ".go") {
			continue
		}

		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if fn, ok := node.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				for _, nd := range fn.Body.List {
					if expr, ok := nd.(*ast.ExprStmt); ok {
						if x, ok := expr.X.(*ast.CallExpr); ok && isExit(x) {
							pass.Reportf(x.Pos(), "os.Exit call in main package")

							break
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}
