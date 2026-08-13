package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// bmiCmd 计算 BMI 并给出健康建议。
var bmiCmd = &cobra.Command{
	Use:   "bmi <身高cm> <体重kg>",
	Short: "计算 BMI 并给出健康建议",
	Long: `根据身高（厘米）与体重（千克）计算 BMI 指数，并给出参考分类。

BMI = 体重(kg) / 身高(m)²
分类（WHO 成人标准）：
  < 18.5        偏瘦
  18.5 ~ 24.9   正常
  25.0 ~ 29.9   超重
  >= 30.0       肥胖

示例:
  philokun bmi 175 68
  philokun bmi 160 55 -u        # 同时显示理想体重区间`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		h, err := strconv.ParseFloat(strings.TrimSpace(args[0]), 64)
		if err != nil || h <= 0 {
			return fmt.Errorf("身高必须是正数（厘米）: %q", args[0])
		}
		w, err := strconv.ParseFloat(strings.TrimSpace(args[1]), 64)
		if err != nil || w <= 0 {
			return fmt.Errorf("体重必须是正数（千克）: %q", args[1])
		}
		showIdeal, _ := cmd.Flags().GetBool("ideal")

		m := h / 100
		bmi := w / (m * m)
		fmt.Printf("身高 %.0f cm，体重 %.1f kg\n", h, w)
		fmt.Printf("BMI = %.1f\n", bmi)
		fmt.Printf("分类：%s\n", bmiCategory(bmi))
		if showIdeal {
			lo, hi := idealWeightRange(m)
			fmt.Printf("理想体重区间（18.5~24.9）：%.1f ~ %.1f kg\n", lo, hi)
		}
		return nil
	},
}

// bmiCategory 返回 BMI 分类文字。
func bmiCategory(bmi float64) string {
	switch {
	case bmi < 18.5:
		return "偏瘦"
	case bmi < 25:
		return "正常"
	case bmi < 30:
		return "超重"
	default:
		return "肥胖"
	}
}

// idealWeightRange 返回身高(m)对应的理想体重区间。
func idealWeightRange(m float64) (float64, float64) {
	return 18.5 * m * m, 24.9 * m * m
}

func init() {
	bmiCmd.Flags().BoolP("ideal", "u", false, "显示理想体重区间")
	rootCmd.AddCommand(bmiCmd)
}
