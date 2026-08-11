package cmd

import (
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQrcodeHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/qrcode?text=hello", nil)
	rec := httptest.NewRecorder()

	qrcodeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("期望 image/png，实际 %q", ct)
	}
	if _, err := png.Decode(rec.Body); err != nil {
		t.Fatalf("返回内容不是合法 PNG: %v", err)
	}
}

func TestQrcodeHandlerEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/qrcode?text=", nil)
	rec := httptest.NewRecorder()

	qrcodeHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("空内容期望 204，实际 %d", rec.Code)
	}
}

func TestQrcodeHandlerPostRejected(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/qrcode?text=hi", nil)
	rec := httptest.NewRecorder()

	qrcodeHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST 期望 405，实际 %d", rec.Code)
	}
}
