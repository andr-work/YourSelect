package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetadata(t *testing.T) {
	base, _ := url.Parse("https://shop.example.com/items/1")
	p := parseProduct(strings.NewReader(`<html><head><title>Fallback</title><meta content="OG title" property="og:title"><meta property="og:image" content="/image.jpg"><script type="application/ld+json">{"@graph":[{"@type":"Product","name":"テスト商品","brand":{"name":"BRAND"},"offers":{"price":"1200","priceCurrency":"JPY"}}]}</script></head></html>`), base)
	if p.Title != "テスト商品" || p.Price != "1200" || p.Brand != "BRAND" || p.ImageURL != "https://shop.example.com/image.jpg" {
		t.Fatalf("unexpected metadata: %+v", p)
	}
}
func TestMetadataOGFallback(t *testing.T) {
	base, _ := url.Parse("https://shop.example.com")
	p := parseProduct(strings.NewReader(`<meta name="description" content="説明"><meta property="og:title" content="マグ"><meta property="product:price:amount" content="2500"><script type="application/ld+json">broken</script>`), base)
	if p.Title != "マグ" || p.Description != "説明" || p.Price != "2500" {
		t.Fatalf("unexpected fallback: %+v", p)
	}
}
func TestBlockedAddresses(t *testing.T) {
	for _, raw := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "::1", "192.168.0.1", "::ffff:127.0.0.1", "0.0.0.0"} {
		if publicIP(net.ParseIP(raw)) {
			t.Errorf("allowed %s", raw)
		}
	}
	if !publicIP(net.ParseIP("8.8.8.8")) {
		t.Error("public ip blocked")
	}
}
func TestCRUDAndPersistence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := filepath.Join(t.TempDir(), "test.db")
	db, e := openDB(path)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	r := router(db)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w
	}
	if w := request("GET", "/api/products", ""); w.Code != 200 || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatal(w.Body.String())
	}
	body := `{"title":"Original","source_url":"https://example.com/item","price":"1200","currency":"JPY"}`
	w := request("POST", "/api/products", body)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var p Product
	json.Unmarshal(w.Body.Bytes(), &p)
	if p.ID != 1 {
		t.Fatal(p)
	}
	if w = request("POST", "/api/products", body); w.Code != 409 {
		t.Fatal("duplicate accepted", w.Code)
	}
	if w = request("PUT", "/api/products/1", strings.Replace(body, "Original", "Updated", 1)); w.Code != 200 {
		t.Fatal(w.Code)
	}
	db.Close()
	db, e = openDB(path)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	r = router(db)
	w = request("GET", "/api/products/1", "")
	json.Unmarshal(w.Body.Bytes(), &p)
	if p.Title != "Updated" {
		t.Fatal("persistence failed", p)
	}
	if w = request("DELETE", "/api/products/1", ""); w.Code != 204 {
		t.Fatal(w.Code)
	}
	if w = request("GET", "/api/products/1", ""); w.Code != 404 {
		t.Fatal(w.Code)
	}
	if w = request("POST", "/api/products", `{"title":"Bad","source_url":"javascript:alert(1)"}`); w.Code != 400 {
		t.Fatal("unsafe URL accepted")
	}
	if w = request("POST", "/api/preview", `{"url":"http://127.0.0.1/private"}`); w.Code != 422 {
		t.Fatal("private fetch allowed", w.Code)
	}
}

func TestDemoNotResurrected(t *testing.T) {
	db, err := openDB(filepath.Join(t.TempDir(), "demo.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = seedDemo(db); err != nil {
		t.Fatal(err)
	}
	var n int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&n)
	if n != 6 {
		t.Fatalf("seed count %d", n)
	}
	db.Exec("DELETE FROM products")
	if err = seedDemo(db); err != nil {
		t.Fatal(err)
	}
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&n)
	if n != 0 {
		t.Fatal("deleted products resurrected")
	}
}
