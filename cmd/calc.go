package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

// calcCmd 一个简单的四则运算计算器，支持 + - * / ^(幂) 与括号。
// 表达式求值采用经典的「双栈」算法（调度场 + 求值），仅用标准库。
var calcCmd = &cobra.Command{
	Use:   "calc <表达式>",
	Short: "计算数学表达式（支持 + - * / ^ 与括号）",
	Long: `计算一个数学表达式并打印结果。

示例:
  philokun calc "1 + 2 * 3"
  philokun calc "(2+3)^2 / 5"
  philokun calc "sqrt(16) + abs(-3)"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expr := strings.ReplaceAll(args[0], " ", "")
		if expr == "" {
			return fmt.Errorf("表达式不能为空")
		}
		v, err := evalExpr(expr)
		if err != nil {
			return err
		}
		// 整数结果不显示小数点
		if v == float64(int64(v)) {
			fmt.Printf("%d\n", int64(v))
		} else {
			fmt.Printf("%.6g\n", v)
		}
		return nil
	},
}

// tokenize 把表达式拆成数字/运算符/括号 token。
func tokenize(s string) ([]string, error) {
	var toks []string
	i := 0
	for i < len(s) {
		c := rune(s[i])
		switch {
		case unicode.IsDigit(c) || c == '.':
			j := i
			dot := 0
			for j < len(s) && (unicode.IsDigit(rune(s[j])) || s[j] == '.') {
				if s[j] == '.' {
					dot++
				}
				j++
			}
			if dot > 1 {
				return nil, fmt.Errorf("非法数字: %s", s[i:j])
			}
			toks = append(toks, s[i:j])
			i = j
		case strings.ContainsRune("+-*/^()", c):
			// 一元负号：表达式开头或运算符后的 '-' 视为负号
			if c == '-' && (len(toks) == 0 || isOp(toks[len(toks)-1]) || toks[len(toks)-1] == "(") {
				// 把负号吸收进下一个数字
				j := i + 1
				for j < len(s) && (unicode.IsDigit(rune(s[j])) || s[j] == '.') {
					j++
				}
				if j > i+1 {
					toks = append(toks, s[i:j])
					i = j
					continue
				}
			}
			toks = append(toks, string(c))
			i++
		case unicode.IsLetter(c):
			// 函数名：sqrt/abs
			j := i
			for j < len(s) && unicode.IsLetter(rune(s[j])) {
				j++
			}
			toks = append(toks, s[i:j])
			i = j
		default:
			return nil, fmt.Errorf("非法字符: %q", string(c))
		}
	}
	return toks, nil
}

func isOp(t string) bool {
	return t == "+" || t == "-" || t == "*" || t == "/" || t == "^"
}

func prec(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	case "^":
		return 3
	}
	return 0
}

// apply 计算一次二元/一元运算。
func apply(op string, a, b float64) (float64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("除以零")
		}
		return a / b, nil
	case "^":
		return pow(a, b), nil
	}
	return 0, fmt.Errorf("未知运算符: %s", op)
}

func pow(a, b float64) float64 {
	r := 1.0
	for i := 0; i < int(b); i++ {
		r *= a
	}
	return r
}

// evalExpr 用双栈法求值：支持 + - * / ^ 与括号，以及 sqrt/abs 函数。
func evalExpr(s string) (float64, error) {
	toks, err := tokenize(s)
	if err != nil {
		return 0, err
	}
	var vals []float64
	var ops []string

	popOp := func() error {
		op := ops[len(ops)-1]
		ops = ops[:len(ops)-1]
		if op == "sqrt" || op == "abs" {
			if len(vals) < 1 {
				return fmt.Errorf("函数 %s 缺少操作数", op)
			}
			a := vals[len(vals)-1]
			vals = vals[:len(vals)-1]
			if op == "sqrt" {
				if a < 0 {
					return fmt.Errorf("sqrt 负数")
				}
				vals = append(vals, sqrt(a))
			} else {
				if a < 0 {
					a = -a
				}
				vals = append(vals, a)
			}
			return nil
		}
		if len(vals) < 2 {
			return fmt.Errorf("运算符 %s 缺少操作数", op)
		}
		b := vals[len(vals)-1]
		a := vals[len(vals)-2]
		vals = vals[:len(vals)-2]
		r, err := apply(op, a, b)
		if err != nil {
			return err
		}
		vals = append(vals, r)
		return nil
	}

	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch {
		case t == "(":
			ops = append(ops, t)
		case t == ")":
			for len(ops) > 0 && ops[len(ops)-1] != "(" {
				if err := popOp(); err != nil {
					return 0, err
				}
			}
			if len(ops) == 0 {
				return 0, fmt.Errorf("括号不匹配")
			}
			ops = ops[:len(ops)-1] // 弹出 '('
		case t == "sqrt" || t == "abs":
			ops = append(ops, t)
		case isOp(t):
			for len(ops) > 0 && ops[len(ops)-1] != "(" && prec(ops[len(ops)-1]) >= prec(t) {
				if err := popOp(); err != nil {
					return 0, err
				}
			}
			ops = append(ops, t)
		default:
			n, err := strconv.ParseFloat(t, 64)
			if err != nil {
				return 0, fmt.Errorf("无法解析数字: %s", t)
			}
			vals = append(vals, n)
		}
	}
	for len(ops) > 0 {
		if ops[len(ops)-1] == "(" {
			return 0, fmt.Errorf("括号不匹配")
		}
		if err := popOp(); err != nil {
			return 0, err
		}
	}
	if len(vals) != 1 {
		return 0, fmt.Errorf("表达式非法")
	}
	return vals[0], nil
}

func sqrt(a float64) float64 {
	// 牛顿迭代法求平方根（避免引入 math 依赖，保持轻量）
	x := a
	if x == 0 {
		return 0
	}
	for i := 0; i < 50; i++ {
		x = (x + a/x) / 2
	}
	return x
}

func init() {
	rootCmd.AddCommand(calcCmd)
}
