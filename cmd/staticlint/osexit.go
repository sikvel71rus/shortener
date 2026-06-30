package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var osExitAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "forbids direct os.Exit calls in func main of package main",
	Run:  runOSExitCheck,
}

func runOSExitCheck(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if isOSExitCall(pass, call) {
					pass.Reportf(call.Pos(), "direct os.Exit call is forbidden in func main of package main")
				}
				return true
			})
		}
	}

	return nil, nil
}

func isOSExitCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Exit" {
		return false
	}

	fn, ok := pass.TypesInfo.Uses[selector.Sel].(*types.Func)
	if !ok {
		return false
	}

	return fn.Pkg() != nil && fn.Pkg().Path() == "os" && fn.Name() == "Exit"
}
