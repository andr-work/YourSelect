# yourselect

「好きが、ショップになる。」URLで集める、自分だけのオンラインセレクトショップのPoCです。ユーザーは **ドレ** 固定で、認証・決済機能はありません。

## 起動

Node.js 22.13以上、Go 1.25以上が必要です。SQLiteはPure Goドライバを使用するため別途インストール不要です。

```bash
npm ci
npm run dev
```

http://localhost:3000 を開いてください。Next.jsとGin APIが同時に起動します。GoがPATHにない場合は `GO_BIN=/path/to/go npm run dev` と指定できます。

初期状態は空のショップです。写真付きサンプル6件で確認する場合は、初回に `SEED_DEMO=1 npm run dev` で起動してください。サンプルの画像・価格・リンクはデモ用です。

### 個別起動

```bash
# ターミナル1
cd backend
go run .

# ターミナル2（プロジェクトルート）
npm run dev:web
```

- SQLite保存先：`backend/data/yourselect.db`（個別起動時はbackendを作業ディレクトリにする）
- API：`http://127.0.0.1:8080`
- 環境変数：`DATABASE_PATH`（DBパス）、`API_ADDR`（API待受）、`API_ORIGIN`（Next.js転送先）
- 同じURLの二重登録は拒否します。商品情報は保存時のスナップショットで、自動同期はしません。

## できること

- URLから商品名、画像、説明、ブランド、価格、通貨を取得（OGP / Twitter Card / JSON-LD Product）。
- 取得内容を確認して登録。取得が失敗した場合や情報が不足する場合は手入力で補完。
- 商品一覧、検索、カテゴリ絞り込み、新しい順・古い順の切替。
- 商品詳細と元の販売ページへのリンク。
- 商品管理ページ、編集、確認ダイアログ付き削除。
- SQLiteへの永続保存、空状態・読込中・エラー・画像読込失敗時の表示。
- 外部商品画像をGoでリサイズ・JPEG変換し、一覧800×1000・管理160×160・詳細1200×1500で表示。商品全体が見えるよう中央配置し、白い余白を追加。既存の商品にも表示時に適用。
- Astryxのclassic-gallery / detail-pageテンプレートを元にしたレスポンシブUI。

JavaScript実行後にしか表示されない情報、ログインが必要なサイト、Botアクセスを拒否するサイトは自動取得できない場合があります。価格・在庫は販売サイトで確認してください。取得APIはHTTP(S)の80/443ポートだけに接続し、リダイレクト時もプライベートIPを遮断します。認証のないローカル検証用PoCのため、インターネットへの公開を前提としていません。

変換画像は `backend/data/images/` にキャッシュします（`IMAGE_CACHE_DIR` で変更可能）。元URLは保持され、キャッシュを削除すると次回表示時に再生成します。JPEG・PNG・GIF（先頭フレーム）・WebPに対応し、JPEGのEXIF回転も反映します。取得・変換が失敗した画像は同じサイズの代替表示になります。

## 検証

```bash
npm run typecheck
npm run build
npx astryx doctor --built .
cd backend
go test ./...
go vet ./...
```

## API

| メソッド | パス | 機能 |
|---|---|---|
| GET | /api/images?url=...&variant=collection | 画像変換・キャッシュ（thumbnail / detailも指定可） |
| GET | /api/health | ヘルスチェック |
| GET | /api/products | 商品一覧 |
| POST | /api/preview | URL情報取得（`{"url":"https://..."}`） |
| POST | /api/products | 商品登録 |
| GET | /api/products/:id | 商品詳細 |
| PUT | /api/products/:id | 商品編集 |
| DELETE | /api/products/:id | 商品削除 |
