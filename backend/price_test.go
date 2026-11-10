package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizePrice(t *testing.T) {
	for _, tt := range []struct{ raw, currency, want string }{
		{"560.00\u00a0EUR", "EUR", "560.00"},
		{"560.00 EUR", "EUR", "560.00"},
		{" EUR 560.00 ", "EUR", "560.00"},
		{"€560.00", "EUR", "560.00"},
		{"560,00 €", "EUR", "560.00"},
		{"1.234,56 EUR", "EUR", "1234.56"},
		{"1\u202f234,56 EUR", "EUR", "1234.56"},
		{"1,234.56 USD", "USD", "1234.56"},
		{"￥12,000", "JPY", "12000"},
		{"1200", "JPY", "1200"},
		{"0", "JPY", "0"},
		{"", "JPY", ""},
	} {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := normalizePrice(tt.raw, tt.currency)
			if err != nil || got != tt.want {
				t.Fatalf("normalizePrice(%q,%q) = %q, %v; want %q", tt.raw, tt.currency, got, err, tt.want)
			}
		})
	}
	for _, s := range []string{"-560.00 EUR", "NaN", "Inf", "100–200 EUR", "560.00 USD", "from 560 EUR", "1,2,3", "5 60 EUR"} {
		if got, err := normalizePrice(s, "EUR"); err == nil {
			t.Errorf("accepted %q as %q", s, got)
		}
	}
}

func TestOurLegacyExtractAndSave(t *testing.T) {
	body, err := os.Open("testdata/ourlegacy.html")
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	base, _ := url.Parse("https://www.ourlegacy.com/mini-jacket-black-worsted-wool")
	p := parseProduct(body, base)
	if p.Price != "560.00" || p.Currency != "EUR" || p.Title != "Mini Jacket - Black Worsted Wool" || p.Brand != "Our Legacy" || p.ImageURL == "" {
		t.Fatalf("unexpected product: %+v", p)
	}
	db, err := openDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := router(db)
	for i, price := range []string{p.Price, "560.00\u00a0EUR"} {
		p.Price = price
		p.SourceURL = base.String() + strings.Repeat("/", i)
		data, _ := json.Marshal(p)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/products", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 201 {
			t.Fatalf("save %q: %d %s", price, w.Code, w.Body.String())
		}
		var saved Product
		if err = json.Unmarshal(w.Body.Bytes(), &saved); err != nil {
			t.Fatal(err)
		}
		if saved.Price != "560.00" || saved.Currency != "EUR" {
			t.Fatalf("unexpected saved price: %+v", saved)
		}
		var storedPrice, storedCurrency string
		if err = db.QueryRow("SELECT price,currency FROM products WHERE id=?", saved.ID).Scan(&storedPrice, &storedCurrency); err != nil {
			t.Fatal(err)
		}
		if storedPrice != "560.00" || storedCurrency != "EUR" {
			t.Fatalf("stored %q %q", storedPrice, storedCurrency)
		}
	}
}

func TestUnsupportedMetadataPrice(t *testing.T) {
	base, _ := url.Parse("https://example.com/product")
	p := parseProduct(strings.NewReader(`<meta property="og:title" content="Product"><meta property="product:price:amount" content="Contact for price">`), base)
	if p.Price != "" {
		t.Fatalf("unusable price returned: %q", p.Price)
	}
	if err := validate(&p); err != nil {
		t.Fatalf("missing price should not block saving: %v", err)
	}
}
