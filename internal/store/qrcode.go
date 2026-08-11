package store

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCode 把任意文本/网址/数字编码成 PNG 格式的二维码字节流。
// 它使用 Medium 纠错级别，兼容主流扫码 App，并自动根据内容长度选择合适的版本。
func GenerateQRCode(content string) ([]byte, error) {
	if content == "" {
		return nil, fmt.Errorf("内容不能为空")
	}
	// Medium 级别：约 15% 的容错，兼顾可读性与鲁棒性。
	png, err := qrcode.Encode(content, qrcode.Medium, 512)
	if err != nil {
		return nil, fmt.Errorf("生成二维码失败: %w", err)
	}
	return png, nil
}
