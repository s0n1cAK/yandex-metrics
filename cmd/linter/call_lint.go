package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func runCallLinter(pass *analysis.Pass, n ast.Node) bool {
	// Проверяем, является ли узел вызовом функции
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return true
	}

	// Проверяем, что вызывается селектор (например, log.Fatal, os.Exit)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return true
	}

	// Получаем объект метода через Uses
	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil {
		return true
	}

	// Проверяем, что это функция из нужного пакета
	if pkg := obj.Pkg(); pkg != nil {
		pkgPath := pkg.Path()
		methodName := obj.Name()

		// Проверяем конкретные вызовы log.Fatal, log.Fatalf, log.Fatalln, os.Exit
		if (pkgPath == "log" && methodName == "Fatal") ||
			(pkgPath == "log" && methodName == "Fatalf") ||
			(pkgPath == "log" && methodName == "Fatalln") ||
			(pkgPath == "os" && methodName == "Exit") {
			// Проверяем, находимся ли мы внутри функции main пакета main
			if !isInsideMainFunction(pass, n) {
				pass.Reportf(n.Pos(), "использование %s.%s запрещено вне функции main пакета main", pkgPath, methodName)
			}
		}
	}

	return true
}

// isInsideMainFunction проверяет, находится ли узел внутри функции main пакета main
func isInsideMainFunction(pass *analysis.Pass, node ast.Node) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}

	nodePos := node.Pos()
	found := false

	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if found {
				return false
			}

			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok || funcDecl.Name == nil || funcDecl.Name.Name != "main" || funcDecl.Body == nil {
				return true
			}

			if nodePos >= funcDecl.Body.Pos() && nodePos <= funcDecl.Body.End() {
				found = true
			}

			return false
		})

		if found {
			break
		}
	}

	return found
}
