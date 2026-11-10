export type Product = {
  id: number;
  title: string;
  description: string;
  image_url: string;
  price: string;
  currency: string;
  brand: string;
  category: string;
  note: string;
  source_url: string;
  created_at: string;
};
export const categories = [
  "すべて",
  "インテリア",
  "ファッション",
  "暮らしの道具",
  "アート・本",
  "その他",
];
export const blankProduct: Product = {
  id: 0,
  title: "",
  description: "",
  image_url: "",
  price: "",
  currency: "JPY",
  brand: "",
  category: "その他",
  note: "",
  source_url: "",
  created_at: "",
};
export async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });
  const body = await res
    .json()
    .catch(() => ({ error: "サーバーに接続できませんでした。" }));
  if (!res.ok) throw new Error(body.error || "処理に失敗しました。");
  return body;
}
export function priceLabel(p: Product) {
  if (!p.price) return "価格は販売サイトで確認";
  const n = Number(p.price.replaceAll(",", ""));
  if (!Number.isFinite(n)) return p.price;
  try {
    return new Intl.NumberFormat("ja-JP", {
      style: "currency",
      currency: p.currency || "JPY",
      maximumFractionDigits: p.currency === "JPY" ? 0 : 2,
    }).format(n);
  } catch {
    return `${p.price} ${p.currency}`;
  }
}
export function sourceLabel(url: string) {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return "";
  }
}
