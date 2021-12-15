package envhunter

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"strconv"

	"github.com/wayneashleyberry/envhunter/pkg/config"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "envhunter",
		Doc:      fmt.Sprintf("ensures referenced environment variables are described in an %s file", config.ExampleFile),
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.ExprStmt)(nil),
		(*ast.AssignStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		var call *ast.CallExpr

		if assign, ok := n.(*ast.AssignStmt); ok {
			if assigncall, ok := assign.Rhs[0].(*ast.CallExpr); ok {
				call = assigncall
			}
		}

		if expr, ok := n.(*ast.ExprStmt); ok {
			if exprcall, ok := expr.X.(*ast.CallExpr); ok {
				call = exprcall
			}
		}

		if call == nil {
			return
		}

		if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel, ok := fun.X.(*ast.Ident); ok {
				packageName := sel.Name
				funcName := fun.Sel.Name

				// Report on calls to os.Getenv
				if packageName == "os" && funcName == "Getenv" {
					if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
						key, err := strconv.Unquote(basicLit.Value)
						if err != nil {
							panic(err)
						}

						pass.Reportf(basicLit.Pos(), "found environment variable: '%s'", key)
					}
				}

				// Report on calls to envconfig.Process and envconfig.MustProcess
				if packageName == "envconfig" && (funcName == "Process" || funcName == "MustProcess") {
					if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
						prefix, err := strconv.Unquote(basicLit.Value)
						if err != nil {
							panic(err)
						}

						var ident *ast.Ident

						if x, ok := call.Args[1].(*ast.Ident); ok {
							ident = x
						}

						if unary, ok := call.Args[1].(*ast.UnaryExpr); ok {
							if x, ok := unary.X.(*ast.Ident); ok {
								ident = x
							}
						}

						if ident == nil {
							return
						}

						var ident2 *ast.Ident

						if valueSpec, ok := ident.Obj.Decl.(*ast.ValueSpec); ok {
							if i, ok := valueSpec.Type.(*ast.Ident); ok {
								ident2 = i
							}
						}

						if assignStmt, ok := ident.Obj.Decl.(*ast.AssignStmt); ok {
							if unaryExpr, ok := assignStmt.Rhs[0].(*ast.UnaryExpr); ok {
								if comp, ok := unaryExpr.X.(*ast.CompositeLit); ok {
									if ii, ok := comp.Type.(*ast.Ident); ok {
										ident2 = ii
									}
								}
							}
						}

						if ident2 == nil {
							return
						}

						fields := getFieldsFromObject(ident2.Obj)

						for _, f := range fields {
							name := f.Names[0].Name

							computedKey := strings.ToUpper(name)

							if f.Tag != nil && f.Tag.Value != "" {
								tag := reflect.StructTag(f.Tag.Value[1 : len(f.Tag.Value)-1])
								if override, ok := tag.Lookup("envconfig"); ok {
									computedKey = strings.ToUpper(override)
								}
							}

							if prefix != "" {
								computedKey = strings.ToUpper(prefix) + "_" + computedKey
							}

							// TODO: it might be more useful to report the position of the field, as opposed to the caller?
							pass.Reportf(basicLit.Pos(), "found computed environment variable: '%s'", computedKey)
						}
					}
				}
			}
		}
	})

	return nil, nil
}

func getFieldsFromObject(obj *ast.Object) []*ast.Field {
	fields := []*ast.Field{}

	if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
		if st, ok := ts.Type.(*ast.StructType); ok {
			if st.Fields.NumFields() == 0 {
				return fields
			}

			for _, f := range st.Fields.List {
				if len(f.Names) == 1 {
					name := f.Names[0].Name

					// Ignore private fields
					if strings.ToLower(name[0:1]) == name[0:1] {
						continue
					}

					fields = append(fields, f)

					continue
				}

				if len(f.Names) == 0 {
					if i, ok := f.Type.(*ast.Ident); ok {
						fields = append(fields, getFieldsFromObject(i.Obj)...)
					}
				}
			}
		}
	}

	return fields
}
