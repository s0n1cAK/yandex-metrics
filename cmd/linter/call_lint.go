package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func runCallLinter(pass *analysis.Pass, n ast.Node) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return true
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return true
	}

	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil {
		return true
	}

	if pkg := obj.Pkg(); pkg != nil {
		pkgPath := pkg.Path()
		methodName := obj.Name()

		if (pkgPath == "log" && methodName == "Fatal") ||
			(pkgPath == "log" && methodName == "Fatalf") ||
			(pkgPath == "log" && methodName == "Fatalln") ||
			(pkgPath == "os" && methodName == "Exit") {
			if !isInsideMainFunction(pass, n) {
				pass.Reportf(n.Pos(), "использование %s.%s запрещено вне функции main пакета main", pkgPath, methodName)
			}
		}
	}

	return true
}

func isInsideMainFunction(pass *analysis.Pass, node ast.Node) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}

	nodePos := node.Pos()
	found := false

	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if funcDecl.Name == nil || funcDecl.Name.Name != "main" {
				return true
			}

			if funcDecl.Body == nil {
				return true
			}

			bodyStart := funcDecl.Body.Pos()
			bodyEnd := funcDecl.Body.End()
			if nodePos >= bodyStart && nodePos <= bodyEnd {
				found = true
				return false
			}
			return true
		})
		if found {
			break
		}
	}

	return found
}
