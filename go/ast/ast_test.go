package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/UB-IAD/sd-ai/go/ast"
)

func TestExprAppendEquation(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expr
		expected string
	}{
		// Const tests
		{
			name:     "simple integer",
			expr:     &ast.Const{Value: 42},
			expected: "42",
		},
		{
			name:     "floating point",
			expr:     &ast.Const{Value: 3.14159},
			expected: "3.14159",
		},
		{
			name:     "negative number",
			expr:     &ast.Const{Value: -10.5},
			expected: "-10.5",
		},
		{
			name:     "scientific notation",
			expr:     &ast.Const{Value: 1.23e-4},
			expected: "0.000123",
		},

		// Var tests
		{
			name:     "simple variable",
			expr:     &ast.Var{Ident: "x"},
			expected: "x",
		},
		{
			name:     "variable with underscores",
			expr:     &ast.Var{Ident: "my_variable_name"},
			expected: "my_variable_name",
		},

		// Call tests
		{
			name:     "function call without args",
			expr:     &ast.Call{Fn: "rand", Args: []ast.Expr{}},
			expected: "rand()",
		},
		{
			name: "function call with one arg",
			expr: &ast.Call{
				Fn:   "sin",
				Args: []ast.Expr{&ast.Var{Ident: "x"}},
			},
			expected: "sin(x)",
		},
		{
			name: "function call with multiple args",
			expr: &ast.Call{
				Fn: "max",
				Args: []ast.Expr{
					&ast.Var{Ident: "a"},
					&ast.Var{Ident: "b"},
					&ast.Const{Value: 10},
				},
			},
			expected: "max(a, b, 10)",
		},
		{
			name: "nested function calls",
			expr: &ast.Call{
				Fn: "pow",
				Args: []ast.Expr{
					&ast.Call{
						Fn:   "sin",
						Args: []ast.Expr{&ast.Var{Ident: "x"}},
					},
					&ast.Const{Value: 2},
				},
			},
			expected: "pow(sin(x), 2)",
		},

		// UnaryOp tests
		{
			name: "unary minus with variable",
			expr: &ast.UnaryOp{
				Op:   "-",
				Expr: &ast.Var{Ident: "x"},
			},
			expected: "-x",
		},
		{
			name: "unary plus with constant",
			expr: &ast.UnaryOp{
				Op:   "+",
				Expr: &ast.Const{Value: 5},
			},
			expected: "+5",
		},
		{
			name: "unary minus with binary expression (needs parens)",
			expr: &ast.UnaryOp{
				Op: "-",
				Expr: &ast.BinaryOp{
					Op:  "+",
					Lhs: &ast.Var{Ident: "a"},
					Rhs: &ast.Var{Ident: "b"},
				},
			},
			expected: "-(a + b)",
		},

		// BinaryOp tests - basic operations
		{
			name: "simple addition",
			expr: &ast.BinaryOp{
				Op:  "+",
				Lhs: &ast.Var{Ident: "a"},
				Rhs: &ast.Var{Ident: "b"},
			},
			expected: "a + b",
		},
		{
			name: "simple subtraction",
			expr: &ast.BinaryOp{
				Op:  "-",
				Lhs: &ast.Var{Ident: "x"},
				Rhs: &ast.Const{Value: 10},
			},
			expected: "x - 10",
		},
		{
			name: "simple multiplication",
			expr: &ast.BinaryOp{
				Op:  "*",
				Lhs: &ast.Const{Value: 2},
				Rhs: &ast.Var{Ident: "y"},
			},
			expected: "2 * y",
		},
		{
			name: "simple division",
			expr: &ast.BinaryOp{
				Op:  "/",
				Lhs: &ast.Var{Ident: "total"},
				Rhs: &ast.Var{Ident: "count"},
			},
			expected: "total / count",
		},

		// BinaryOp tests - precedence with parentheses
		{
			name: "multiplication of addition (needs parens on left)",
			expr: &ast.BinaryOp{
				Op: "*",
				Lhs: &ast.BinaryOp{
					Op:  "+",
					Lhs: &ast.Var{Ident: "a"},
					Rhs: &ast.Var{Ident: "b"},
				},
				Rhs: &ast.Var{Ident: "c"},
			},
			expected: "(a + b) * c",
		},
		{
			name: "addition of multiplication (no parens needed)",
			expr: &ast.BinaryOp{
				Op: "+",
				Lhs: &ast.BinaryOp{
					Op:  "*",
					Lhs: &ast.Var{Ident: "a"},
					Rhs: &ast.Var{Ident: "b"},
				},
				Rhs: &ast.Var{Ident: "c"},
			},
			expected: "a * b + c",
		},
		{
			name: "subtraction chain (right-associative, needs parens)",
			expr: &ast.BinaryOp{
				Op:  "-",
				Lhs: &ast.Var{Ident: "a"},
				Rhs: &ast.BinaryOp{
					Op:  "-",
					Lhs: &ast.Var{Ident: "b"},
					Rhs: &ast.Var{Ident: "c"},
				},
			},
			expected: "a - (b - c)",
		},
		{
			name: "division chain (right-associative, needs parens)",
			expr: &ast.BinaryOp{
				Op:  "/",
				Lhs: &ast.Var{Ident: "a"},
				Rhs: &ast.BinaryOp{
					Op:  "/",
					Lhs: &ast.Var{Ident: "b"},
					Rhs: &ast.Var{Ident: "c"},
				},
			},
			expected: "a / (b / c)",
		},
		{
			name: "complex expression with multiple precedence levels",
			expr: &ast.BinaryOp{
				Op: "+",
				Lhs: &ast.BinaryOp{
					Op:  "*",
					Lhs: &ast.Var{Ident: "a"},
					Rhs: &ast.Var{Ident: "b"},
				},
				Rhs: &ast.BinaryOp{
					Op:  "/",
					Lhs: &ast.Var{Ident: "c"},
					Rhs: &ast.BinaryOp{
						Op:  "+",
						Lhs: &ast.Var{Ident: "d"},
						Rhs: &ast.Var{Ident: "e"},
					},
				},
			},
			expected: "a * b + c / (d + e)",
		},

		// If tests
		{
			name: "simple if expression",
			expr: &ast.If{
				Cond: &ast.Var{Ident: "condition"},
				Then: &ast.Const{Value: 1},
				Else: &ast.Const{Value: 0},
			},
			expected: "if condition then 1 else 0",
		},
		{
			name: "if with complex condition",
			expr: &ast.If{
				Cond: &ast.BinaryOp{
					Op:  ">",
					Lhs: &ast.Var{Ident: "x"},
					Rhs: &ast.Const{Value: 0},
				},
				Then: &ast.Var{Ident: "positive"},
				Else: &ast.Var{Ident: "negative"},
			},
			expected: "if x > 0 then positive else negative",
		},
		{
			name: "nested if expressions",
			expr: &ast.If{
				Cond: &ast.Var{Ident: "a"},
				Then: &ast.If{
					Cond: &ast.Var{Ident: "b"},
					Then: &ast.Const{Value: 1},
					Else: &ast.Const{Value: 2},
				},
				Else: &ast.Const{Value: 3},
			},
			expected: "if a then if b then 1 else 2 else 3",
		},

		// Subscript tests
		{
			name: "subscript with single index",
			expr: &ast.Subscript{
				Ident: "array",
				Index: []ast.IExpr{
					&ast.IndexExpr{
						Expr: &ast.Var{Ident: "i"},
					},
				},
			},
			expected: "array[i]",
		},
		{
			name: "subscript with multiple indices",
			expr: &ast.Subscript{
				Ident: "matrix",
				Index: []ast.IExpr{
					&ast.IndexExpr{
						Expr: &ast.Var{Ident: "row"},
					},
					&ast.IndexExpr{
						Expr: &ast.Var{Ident: "col"},
					},
				},
			},
			expected: "matrix[row, col]",
		},
		{
			name: "subscript with wildcard",
			expr: &ast.Subscript{
				Ident: "data",
				Index: []ast.IExpr{
					&ast.Wildcard{},
				},
			},
			expected: "data[*]",
		},
		{
			name: "subscript with star range",
			expr: &ast.Subscript{
				Ident: "values",
				Index: []ast.IExpr{
					&ast.StarRange{Ident: "time"},
				},
			},
			expected: "values[*:time]",
		},
		{
			name: "subscript with range",
			expr: &ast.Subscript{
				Ident: "series",
				Index: []ast.IExpr{
					&ast.Range{
						Start: &ast.Const{Value: 0},
						End:   &ast.Const{Value: 10},
					},
				},
			},
			expected: "series[0:10]",
		},
		{
			name: "subscript with expression index",
			expr: &ast.Subscript{
				Ident: "table",
				Index: []ast.IExpr{
					&ast.IndexExpr{
						Expr: &ast.BinaryOp{
							Op:  "+",
							Lhs: &ast.Var{Ident: "base"},
							Rhs: &ast.Const{Value: 1},
						},
					},
				},
			},
			expected: "table[base + 1]",
		},
		{
			name: "subscript with complex mixed indices",
			expr: &ast.Subscript{
				Ident: "tensor",
				Index: []ast.IExpr{
					&ast.Wildcard{},
					&ast.Range{
						Start: &ast.Var{Ident: "start"},
						End:   &ast.Var{Ident: "end"},
					},
					&ast.StarRange{Ident: "dim"},
				},
			},
			expected: "tensor[*, start:end, *:dim]",
		},

		// Complex combined expressions
		{
			name: "function call with arithmetic expression",
			expr: &ast.Call{
				Fn: "sqrt",
				Args: []ast.Expr{
					&ast.BinaryOp{
						Op: "+",
						Lhs: &ast.BinaryOp{
							Op:  "*",
							Lhs: &ast.Var{Ident: "a"},
							Rhs: &ast.Var{Ident: "a"},
						},
						Rhs: &ast.BinaryOp{
							Op:  "*",
							Lhs: &ast.Var{Ident: "b"},
							Rhs: &ast.Var{Ident: "b"},
						},
					},
				},
			},
			expected: "sqrt(a * a + b * b)",
		},
		{
			name: "subscript in arithmetic expression",
			expr: &ast.BinaryOp{
				Op: "*",
				Lhs: &ast.Subscript{
					Ident: "weights",
					Index: []ast.IExpr{
						&ast.IndexExpr{
							Expr: &ast.Var{Ident: "i"},
						},
					},
				},
				Rhs: &ast.Subscript{
					Ident: "values",
					Index: []ast.IExpr{
						&ast.IndexExpr{
							Expr: &ast.Var{Ident: "i"},
						},
					},
				},
			},
			expected: "weights[i] * values[i]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf []byte
			result := tc.expr.AppendEquation(buf)
			assert.Equal(t, tc.expected, string(result))
		})
	}
}

func TestIExprAppendEquation(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.IExpr
		expected string
	}{
		{
			name:     "wildcard",
			expr:     &ast.Wildcard{},
			expected: "*",
		},
		{
			name:     "star range",
			expr:     &ast.StarRange{Ident: "time"},
			expected: "*:time",
		},
		{
			name: "range with constants",
			expr: &ast.Range{
				Start: &ast.Const{Value: 1},
				End:   &ast.Const{Value: 10},
			},
			expected: "1:10",
		},
		{
			name: "range with variables",
			expr: &ast.Range{
				Start: &ast.Var{Ident: "start"},
				End:   &ast.Var{Ident: "end"},
			},
			expected: "start:end",
		},
		{
			name: "range with expressions",
			expr: &ast.Range{
				Start: &ast.BinaryOp{
					Op:  "-",
					Lhs: &ast.Var{Ident: "n"},
					Rhs: &ast.Const{Value: 1},
				},
				End: &ast.BinaryOp{
					Op:  "+",
					Lhs: &ast.Var{Ident: "n"},
					Rhs: &ast.Const{Value: 1},
				},
			},
			expected: "n - 1:n + 1",
		},
		{
			name: "index expression with simple variable",
			expr: &ast.IndexExpr{
				Expr: &ast.Var{Ident: "idx"},
			},
			expected: "idx",
		},
		{
			name: "index expression with complex expression",
			expr: &ast.IndexExpr{
				Expr: &ast.BinaryOp{
					Op: "*",
					Lhs: &ast.BinaryOp{
						Op:  "+",
						Lhs: &ast.Var{Ident: "row"},
						Rhs: &ast.Const{Value: 1},
					},
					Rhs: &ast.Var{Ident: "cols"},
				},
			},
			expected: "(row + 1) * cols",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf []byte
			result := tc.expr.AppendEquation(buf)
			assert.Equal(t, tc.expected, string(result))
		})
	}
}

func TestAppendEquationWithExistingBuffer(t *testing.T) {
	// Test that AppendEquation correctly appends to an existing buffer
	initial := []byte("result = ")
	expr := &ast.BinaryOp{
		Op:  "+",
		Lhs: &ast.Var{Ident: "x"},
		Rhs: &ast.Const{Value: 42},
	}

	result := expr.AppendEquation(initial)
	assert.Equal(t, "result = x + 42", string(result))

	// Ensure original buffer is not modified if it has capacity
	assert.Equal(t, "result = ", string(initial))
}
