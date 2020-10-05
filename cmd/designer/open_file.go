package main

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/gonutz/wui"
)

func openFile(path string) ([]*wui.Window, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	windows, err := extractWindowsFromCode(string(data))
	w := make([]*wui.Window, len(windows))
	for i := range w {
		w[i] = windows[i].window
	}
	return w, err
}

type windowInCode struct {
	window             *wui.Window
	creationLineNumber int
}

func extractWindowsFromCode(code string) ([]windowInCode, error) {
	var windows []windowInCode
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", code, parser.AllErrors)
	if err != nil {
		return nil, errors.New("Parse error in code: " + err.Error())
	}
	wuiName, err := findWuiPackageImport(f)
	if err != nil {
		return nil, err
	}

	var lastBlock *ast.BlockStmt
	ast.Inspect(f, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			lastBlock = block
		}
		if name, ok := isWuiWindowCreation(f, wuiName, n); ok {
			windows = append(windows, windowInCode{
				window:             buildWindow(name, lastBlock, n),
				creationLineNumber: fset.Position(n.Pos()).Line,
			})
		}
		return true
	})

	return windows, nil
}

func findWuiPackageImport(f *ast.File) (importName string, err error) {
	importName = "wui"
	found := false
	for _, imp := range f.Imports {
		if imp.Path.Value == `"github.com/gonutz/wui"` {
			if found {
				return "", errors.New("wui is imported multiple times")
			}
			found = true
			if imp.Name != nil {
				importName = imp.Name.Name
			}
		}
	}
	if !found {
		return "", errors.New("wui import was not found")
	}
	return importName, nil
}

func isWuiWindowCreation(f *ast.File, wuiName string, n ast.Node) (name string, ok bool) {
	if false {
	} else if assign, ok := n.(*ast.AssignStmt); !ok {
	} else if !(len(assign.Lhs) == 1) {
	} else if variable, ok := assign.Lhs[0].(*ast.Ident); !ok {
	} else if !(variable.Name != "_") {
	} else if !(len(assign.Rhs) == 1) {
	} else if call, ok := assign.Rhs[0].(*ast.CallExpr); !ok {
	} else if !(len(call.Args) == 0) {
	} else if sel, ok := call.Fun.(*ast.SelectorExpr); !ok {
	} else if pkg, ok := sel.X.(*ast.Ident); !ok {
	} else if !(pkg.Name == wuiName) {
	} else if !(sel.Sel.Name == "NewWindow") {
	} else if !containsIdent(f.Unresolved, pkg) {
		// f.Unresolved is usually filled when other package files are parsed as
		// well. In our case we only parse the code in one file and since the
		// wui import is not resolved, because we do not parse wui itself, the
		// package must be in f.Unresolved. If it is not, then the code before
		// wui.NewWindow must have re-defined wui to something else.
	} else {
		return variable.Name, true
	}
	return "", false
}

func containsIdent(ids []*ast.Ident, id *ast.Ident) bool {
	for i := range ids {
		if id == ids[i] {
			return true
		}
	}
	return false
}

// TODO Return error as well. Test first.
func buildWindow(varName string, block *ast.BlockStmt, assignment ast.Node) *wui.Window {
	w := wui.NewWindow()
	first := 0
	for i, stmt := range block.List {
		assign, ok := stmt.(*ast.AssignStmt)
		if ok && assign == assignment {
			first = i
			break
		}
	}
	for _, stmt := range block.List[first+1:] {
		if isReassignment(stmt, varName) {
			break
		}
		if funcName, args, ok := isMethodCallOn(stmt, varName); ok && strings.HasPrefix(funcName, "Set") {
			win := reflect.ValueOf(w)
			method := win.MethodByName(funcName)
			if !method.IsValid() {
				// TODO
				return nil
			}
			values := make([]reflect.Value, len(args))
			for i, arg := range args {
				// TODO Also handle Ident (var/const) instead of only literals.
				// Some more interpretation might be necessary for this.
				lit := arg.(*ast.BasicLit)
				var value reflect.Value
				switch lit.Kind {
				case token.INT:
					n, err := strconv.Atoi(lit.Value)
					if err != nil {
						// TODO Return error.
						panic(err)
					}
					value = reflect.ValueOf(n)
					// TODO Other cases.
				}
				values[i] = value.Convert(method.Type().In(i))
			}
			method.Call(values)
		}
	}
	return w
}

func isReassignment(stmt ast.Stmt, name string) bool {
	if assign, ok := stmt.(*ast.AssignStmt); ok {
		for _, left := range assign.Lhs {
			if id, ok := left.(*ast.Ident); ok && id.Name == name {
				return true
			}
		}
	}
	return false
}

func isMethodCallOn(stmt ast.Stmt, name string) (funcName string, args []ast.Expr, ok bool) {
	if false {
	} else if exprStmt, ok := stmt.(*ast.ExprStmt); !ok {
	} else if call, ok := exprStmt.X.(*ast.CallExpr); !ok {
	} else if sel, ok := call.Fun.(*ast.SelectorExpr); !ok {
	} else if callee, ok := sel.X.(*ast.Ident); !ok {
	} else if !(callee.Name == name) {
	} else {
		return sel.Sel.Name, call.Args, true
	}
	return "", nil, false
}
