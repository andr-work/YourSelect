package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
	_ "modernc.org/sqlite"
)

type Product struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Price       string `json:"price"`
	Currency    string `json:"currency"`
	Brand       string `json:"brand"`
	Category    string `json:"category"`
	Note        string `json:"note"`
	SourceURL   string `json:"source_url"`
	CreatedAt   string `json:"created_at"`
}

const columns = "id,title,description,image_url,price,currency,brand,category,note,source_url,created_at"

func scanProduct(row interface{ Scan(...any) error }) (p Product, err error) {
	err = row.Scan(&p.ID, &p.Title, &p.Description, &p.ImageURL, &p.Price, &p.Currency, &p.Brand, &p.Category, &p.Note, &p.SourceURL, &p.CreatedAt)
	return
}
func validURL(s string) bool {
	u, e := url.Parse(s)
	return e == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "" && u.User == nil
}
func validate(p *Product) error {
	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return errors.New("商品名を入力してください。")
	}
	if len(p.Title) > 500 {
		return errors.New("商品名が長すぎます。")
	}
	if !validURL(p.SourceURL) {
		return errors.New("有効な商品URLを入力してください。")
	}
	if p.ImageURL != "" && !validURL(p.ImageURL) && !(strings.HasPrefix(p.ImageURL, "/") && !strings.HasPrefix(p.ImageURL, "//")) {
		return errors.New("画像URLはhttpまたはhttpsで指定してください。")
	}
	price, err := normalizePrice(p.Price, p.Currency)
	if err != nil {
		return err
	}
	p.Price = price

	if p.Currency == "" {
		p.Currency = "JPY"
	}
	if len(p.Currency) != 3 {
		return errors.New("通貨はJPYなど3文字で入力してください。")
	}
	if p.Category == "" {
		p.Category = "その他"
	}
	return nil
}
func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, e := sql.Open("sqlite", path)
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(1)
	_, e = db.Exec(`PRAGMA busy_timeout=5000; CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY AUTOINCREMENT,title TEXT NOT NULL,description TEXT NOT NULL DEFAULT '',image_url TEXT NOT NULL DEFAULT '',price TEXT NOT NULL DEFAULT '',currency TEXT NOT NULL DEFAULT 'JPY',brand TEXT NOT NULL DEFAULT '',category TEXT NOT NULL DEFAULT 'その他',note TEXT NOT NULL DEFAULT '',source_url TEXT NOT NULL UNIQUE,created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')));`)
	return db, e
}
func router(db *sql.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(func(c *gin.Context) { c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20); c.Next() })
	cacheDir := os.Getenv("IMAGE_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "data/images"
	}
	images := newImageService(cacheDir, safeClient())
	r.GET("/api/images", images.handle)
	r.GET("/api/health", func(c *gin.Context) {
		if e := db.Ping(); e != nil {
			c.JSON(503, gin.H{"error": "DBに接続できません。"})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "user": "ドレ"})
	})
	r.GET("/api/products", func(c *gin.Context) {
		rows, e := db.Query("SELECT " + columns + " FROM products ORDER BY id DESC")
		if e != nil {
			c.JSON(500, gin.H{"error": "商品一覧を取得できませんでした。"})
			return
		}
		defer rows.Close()
		out := []Product{}
		for rows.Next() {
			p, e := scanProduct(rows)
			if e != nil {
				c.JSON(500, gin.H{"error": "商品を読み込めませんでした。"})
				return
			}
			out = append(out, p)
		}
		if rows.Err() != nil {
			c.JSON(500, gin.H{"error": "商品を読み込めませんでした。"})
			return
		}
		c.JSON(200, out)
	})
	r.GET("/api/products/:id", func(c *gin.Context) {
		p, e := scanProduct(db.QueryRow("SELECT "+columns+" FROM products WHERE id=?", c.Param("id")))
		if e != nil {
			c.JSON(404, gin.H{"error": "商品が見つかりません。"})
			return
		}
		c.JSON(200, p)
	})
	r.POST("/api/preview", func(c *gin.Context) {
		var body struct {
			URL string `json:"url"`
		}
		if c.ShouldBindJSON(&body) != nil || !validURL(body.URL) {
			c.JSON(400, gin.H{"error": "httpまたはhttpsの商品URLを入力してください。"})
			return
		}
		p, e := fetchProduct(c.Request.Context(), body.URL)
		if e != nil {
			c.JSON(422, gin.H{"error": "商品情報を取得できませんでした。URLを確認するか、手入力で登録してください。"})
			return
		}
		c.JSON(200, p)
	})
	save := func(c *gin.Context) {
		var p Product
		if c.ShouldBindJSON(&p) != nil {
			c.JSON(400, gin.H{"error": "入力内容を確認してください。"})
			return
		}
		if e := validate(&p); e != nil {
			c.JSON(400, gin.H{"error": e.Error()})
			return
		}
		var result sql.Result
		var e error
		if c.Request.Method == "POST" {
			result, e = db.Exec("INSERT INTO products(title,description,image_url,price,currency,brand,category,note,source_url) VALUES(?,?,?,?,?,?,?,?,?)", p.Title, p.Description, p.ImageURL, p.Price, p.Currency, p.Brand, p.Category, p.Note, p.SourceURL)
		} else {
			result, e = db.Exec("UPDATE products SET title=?,description=?,image_url=?,price=?,currency=?,brand=?,category=?,note=?,source_url=? WHERE id=?", p.Title, p.Description, p.ImageURL, p.Price, p.Currency, p.Brand, p.Category, p.Note, p.SourceURL, c.Param("id"))
		}
		if e != nil {
			if strings.Contains(e.Error(), "UNIQUE") {
				c.JSON(409, gin.H{"error": "このURLの商品はすでに登録されています。"})
			} else {
				c.JSON(500, gin.H{"error": "保存に失敗しました。"})
			}
			return
		}
		if c.Request.Method == "POST" {
			p.ID, _ = result.LastInsertId()
		} else {
			n, _ := result.RowsAffected()
			if n == 0 {
				c.JSON(404, gin.H{"error": "商品が見つかりません。"})
				return
			}
			p.ID, _ = strconv.ParseInt(c.Param("id"), 10, 64)
		}
		p, _ = scanProduct(db.QueryRow("SELECT "+columns+" FROM products WHERE id=?", p.ID))
		status := 200
		if c.Request.Method == "POST" {
			status = 201
		}
		c.JSON(status, p)
	}
	r.POST("/api/products", save)
	r.PUT("/api/products/:id", save)
	r.DELETE("/api/products/:id", func(c *gin.Context) {
		result, e := db.Exec("DELETE FROM products WHERE id=?", c.Param("id"))
		if e != nil {
			c.JSON(500, gin.H{"error": "削除に失敗しました。"})
			return
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			c.JSON(404, gin.H{"error": "商品が見つかりません。"})
			return
		}
		c.Status(204)
	})
	return r
}

// Resolve and validate every dial, including redirects, to block access to local services.
func publicIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsUnspecified() && !ip.Equal(net.ParseIP("169.254.169.254"))
}
func safeClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		if !validURL(req.URL.String()) {
			return errors.New("invalid redirect")
		}
		return nil
	}, Transport: &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, e := net.SplitHostPort(address)
		if e != nil {
			return nil, e
		}
		if port != "80" && port != "443" {
			return nil, errors.New("port blocked")
		}
		ips, e := net.DefaultResolver.LookupIPAddr(ctx, host)
		if e != nil {
			return nil, e
		}
		for _, ip := range ips {
			if !publicIP(ip.IP) {
				return nil, errors.New("private address blocked")
			}
		}
		if len(ips) == 0 {
			return nil, errors.New("no address")
		}
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}}}
}
func fetchProduct(ctx context.Context, raw string) (Product, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", raw, nil)
	if e != nil {
		return Product{}, e
	}
	req.Header.Set("User-Agent", "YourSelect/0.1 (Product metadata preview)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	res, e := safeClient().Do(req)
	if e != nil {
		return Product{}, e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return Product{}, fmt.Errorf("http %d", res.StatusCode)
	}
	body, e := io.ReadAll(io.LimitReader(res.Body, 3<<20+1))
	if e != nil || len(body) > 3<<20 {
		return Product{}, errors.New("page too large")
	}
	return parseProduct(strings.NewReader(string(body)), res.Request.URL), nil
}
func parseProduct(reader io.Reader, base *url.URL) Product {
	p := Product{SourceURL: base.String(), Currency: "JPY", Category: "その他"}
	doc, e := html.Parse(reader)
	if e != nil {
		return p
	}
	meta := map[string]string{}
	var title string
	var structured []any
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			attrs := map[string]string{}
			for _, a := range n.Attr {
				attrs[strings.ToLower(a.Key)] = a.Val
			}
			if n.Data == "meta" {
				key := attrs["property"]
				if key == "" {
					key = attrs["name"]
				}
				meta[strings.ToLower(key)] = attrs["content"]
			}
			if n.Data == "title" && n.FirstChild != nil {
				title = n.FirstChild.Data
			}
			if n.Data == "script" && strings.EqualFold(attrs["type"], "application/ld+json") && n.FirstChild != nil {
				var data any
				if json.Unmarshal([]byte(n.FirstChild.Data), &data) == nil {
					structured = append(structured, data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	p.Title = first(meta["og:title"], meta["twitter:title"], title)
	p.Description = first(meta["og:description"], meta["description"], meta["twitter:description"])
	p.ImageURL = first(meta["og:image"], meta["twitter:image"])
	p.Price = first(meta["product:price:amount"], meta["og:price:amount"])
	p.Currency = first(meta["product:price:currency"], meta["og:price:currency"], "JPY")
	p.Brand = meta["og:site_name"]
	var visit func(any)
	visit = func(v any) {
		switch x := v.(type) {
		case []any:
			for _, v := range x {
				visit(v)
			}
		case map[string]any:
			t := fmt.Sprint(x["@type"])
			if t == "Product" || strings.Contains(t, "Product") {
				p.Title = first(str(x["name"]), p.Title)
				p.Description = first(str(x["description"]), p.Description)
				switch im := x["image"].(type) {
				case string:
					p.ImageURL = im
				case []any:
					if len(im) > 0 {
						p.ImageURL = imageValue(im[0])
					}
				case map[string]any:
					p.ImageURL = imageValue(im)
				}
				switch b := x["brand"].(type) {
				case string:
					p.Brand = b
				case map[string]any:
					p.Brand = first(str(b["name"]), p.Brand)
				}
				offer := x["offers"]
				if list, ok := offer.([]any); ok && len(list) > 0 {
					offer = list[0]
				}
				if o, ok := offer.(map[string]any); ok {
					p.Price = first(str(o["price"]), str(o["lowPrice"]), p.Price)
					p.Currency = first(str(o["priceCurrency"]), p.Currency)
				}
			}
			for _, child := range x {
				visit(child)
			}
		}
	}
	for _, v := range structured {
		visit(v)
	}
	if u, e := url.Parse(p.ImageURL); e == nil && p.ImageURL != "" {
		p.ImageURL = base.ResolveReference(u).String()
		if !validURL(p.ImageURL) {
			p.ImageURL = ""
		}
	}
	p.Title = strings.TrimSpace(p.Title)
	// Unsupported metadata must not populate the editor with an unsavable price.
	p.Price, _ = normalizePrice(p.Price, p.Currency)
	return p
}
func str(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return ""
}
func imageValue(v any) string {
	if m, ok := v.(map[string]any); ok {
		return first(str(m["url"]), str(m["contentUrl"]))
	}
	return str(v)
}
func first(s ...string) string {
	for _, v := range s {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
func main() {
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "data/yourselect.db"
	}
	db, e := openDB(path)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if os.Getenv("SEED_DEMO") == "1" {
		if e = seedDemo(db); e != nil {
			log.Fatal(e)
		}
	}
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	srv := &http.Server{Addr: addr, Handler: router(db), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("YourSelect API: http://%s", addr)
	log.Fatal(srv.ListenAndServe())
}
