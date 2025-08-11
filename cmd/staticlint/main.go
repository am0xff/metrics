package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/ast/inspector"

	"honnef.co/go/tools/staticcheck"
)

// NoOsExitAnalyzer запрещает использование прямого вызова os.Exit в функции main пакета main
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "check for os.Exit calls in main function of main package",
	Run:      runNoOsExit,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	// Проверяем только пакет main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Ищем функции
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl := n.(*ast.FuncDecl)

		// Проверяем только функцию main
		if funcDecl.Name.Name != "main" {
			return
		}

		// Проверяем тело функции на наличие вызовов os.Exit
		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			if callExpr, ok := node.(*ast.CallExpr); ok {
				if isOsExitCall(callExpr, pass.TypesInfo) {
					pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function is prohibited")
				}
			}
			return true
		})
	})

	return nil, nil
}

// isOsExitCall проверяет, является ли вызов вызовом os.Exit
func isOsExitCall(call *ast.CallExpr, info *types.Info) bool {
	if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
				// Дополнительно проверяем через types.Info для точности
				if obj := info.Uses[ident]; obj != nil {
					if pkg := obj.Pkg(); pkg != nil && pkg.Path() == "os" {
						return true
					}
				}
			}
		}
	}
	return false
}

func main() {
	// Топ-5 самых популярных стандартных анализаторов
	analyzers := []*analysis.Analyzer{
		printf.Analyzer,       // проверка форматных строк (самый популярный)
		structtag.Analyzer,    // проверка тегов структур
		unreachable.Analyzer,  // проверка недостижимого кода
		loopclosure.Analyzer,  // проверка замыканий в циклах
		unusedresult.Analyzer, // проверка неиспользуемых результатов
	}

	// Добавляем все SA анализаторы из staticcheck
	for _, analyzer := range staticcheck.Analyzers {
		if strings.HasPrefix(analyzer.Analyzer.Name, "SA") {
			analyzers = append(analyzers, analyzer.Analyzer)
		}
	}

	// Добавляем анализаторы других классов staticcheck
	for _, analyzer := range staticcheck.Analyzers {
		name := analyzer.Analyzer.Name
		// ST - стилистические проверки
		if strings.HasPrefix(name, "ST1000") || // package comment
			strings.HasPrefix(name, "ST1003") || // naming conventions
			strings.HasPrefix(name, "ST1016") { // receiver names
			analyzers = append(analyzers, analyzer.Analyzer)
		}
		// S - простые проверки
		if strings.HasPrefix(name, "S1000") || // redundant if statement
			strings.HasPrefix(name, "S1002") || // redundant comparison
			strings.HasPrefix(name, "S1025") { // redundant sprintf
			analyzers = append(analyzers, analyzer.Analyzer)
		}
	}

	// Собственный анализатор
	analyzers = append(analyzers, NoOsExitAnalyzer)

	multichecker.Main(analyzers...)
}
