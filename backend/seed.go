package main

import "database/sql"

// Demo records are opt-in, and are inserted only into a new, empty database.
func seedDemo(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec("CREATE TABLE IF NOT EXISTS demo_seed (done INTEGER NOT NULL)"); err != nil {
		return err
	}
	var seeded int
	if err = tx.QueryRow("SELECT COUNT(*) FROM demo_seed").Scan(&seeded); err != nil {
		return err
	}
	if seeded > 0 {
		return nil
	}
	var n int
	if e := tx.QueryRow("SELECT COUNT(*) FROM products").Scan(&n); e != nil {
		return e
	}
	if n > 0 {
		if _, err = tx.Exec("INSERT INTO demo_seed(done) VALUES(1)"); err != nil {
			return err
		}
		return tx.Commit()
	}
	items := []struct{ title, brand, category, price, image, note string }{
		{"静かな時間をつくる、ラウンジチェア", "SAMPLE / FURNITURE", "インテリア", "48000", "photo-1567538096630-e0c55bd6374c", "余白のある部屋に、ひとつだけ置きたい椅子。曲線と木の表情が好きです。"},
		{"毎日手に取りたい、セラミックマグ", "SAMPLE / CERAMICS", "暮らしの道具", "3300", "photo-1514228742587-6b1558fcca3d", "朝のコーヒーが少し楽しみになる、やさしい佇まい。"},
		{"光をやわらかく包む、テーブルランプ", "SAMPLE / LIGHTING", "インテリア", "19800", "photo-1507473885765-e6ed057f782c", "夜の読書にちょうどいい、あたたかな光。"},
		{"日常に寄り添う、デイパック", "SAMPLE / BAGS", "ファッション", "24200", "photo-1553062407-98eeb64c6a62", "使い込むほど、自分のものになっていく。"},
		{"かたちを楽しむ、フラワーベース", "SAMPLE / OBJECTS", "インテリア", "6600", "photo-1490312278390-ab64016e0aa9", "花を飾らない日も、オブジェとして楽しめるもの。"},
		{"視点を変える、一冊のアートブック", "SAMPLE / BOOKS", "アート・本", "4400", "books", "ページをめくるたび、新しい発見がある。"},
	}
	for i, p := range items {
		_, e := tx.Exec("INSERT INTO products(title,brand,category,price,image_url,note,description,source_url) VALUES(?,?,?,?,?,?,?,?)", p.title, p.brand, p.category, p.price, "/demo/"+p.image+".jpg", p.note, "ショップの雰囲気を確認するためのサンプル商品です。画像・価格はデモ用です。", "https://example.com/products/demo-"+string(rune('1'+i)))
		if e != nil {
			return e
		}
	}
	if _, err = tx.Exec("INSERT INTO demo_seed(done) VALUES(1)"); err != nil {
		return err
	}
	return tx.Commit()
}
