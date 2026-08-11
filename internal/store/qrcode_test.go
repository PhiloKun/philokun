package store

import (
	"bytes"
	"image/png"
	"testing"
)

func TestGenerateQRCodeValid(t *testing.T) {
	pngBytes, err := GenerateQRCode("https://github.com/PhiloKun/philokun")
	if err != nil {
		t.Fatalf("生成二维码失败: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("生成的二维码 PNG 不应为空")
	}

	// 校验确实是合法的 PNG（有 PNG 文件头且能被解码）。
	if !bytes.HasPrefix(pngBytes, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		t.Fatal("输出不是合法的 PNG 文件头")
	}
	if _, err := png.Decode(bytes.NewReader(pngBytes)); err != nil {
		t.Fatalf("PNG 解码失败: %v", err)
	}
}

func TestGenerateQRCodeEmpty(t *testing.T) {
	_, err := GenerateQRCode("")
	if err == nil {
		t.Fatal("空内容应当返回错误")
	}
}

func TestGenerateQRCodeChinese(t *testing.T) {
	// 中文字符是常见场景，确保能正常编码。
	pngBytes, err := GenerateQRCode("philokun 效率工具")
	if err != nil {
		t.Fatalf("中文内容生成失败: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("中文二维码不应为空")
	}
}
