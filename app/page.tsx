"use client";
// Adapted from Astryx classic-gallery: centered intro, filter tabs, responsive media grid.
import { useEffect, useState } from "react";
import NextLink from "next/link";
import {
  VStack,
  HStack,
  Layout,
  LayoutContent,
  StackItem,
} from "@astryxdesign/core/Layout";
import { Link } from "@astryxdesign/core/Link";
import { Center } from "@astryxdesign/core/Center";
import { Text, Heading } from "@astryxdesign/core/Text";
import { Grid } from "@astryxdesign/core/Grid";
import { TabList, Tab } from "@astryxdesign/core/TabList";
import { Button } from "@astryxdesign/core/Button";
import { TextInput } from "@astryxdesign/core/TextInput";
import { Icon } from "@astryxdesign/core/Icon";
import { ArrowUpRight, Plus, ArrowDown, Search, Link2 } from "lucide-react";
import { Shell } from "../components/shell";
import { Editor } from "../components/editor";
import { ProductImage } from "../components/product-image";
import { api, categories, Product, priceLabel } from "../lib/api";
export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [category, setCategory] = useState("すべて");
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState("new");
  const [editor, setEditor] = useState(false);
  const [editProduct, setEditProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [manage, setManage] = useState(false);
  async function load() {
    setLoading(true);
    setError("");
    try {
      setProducts(await api<Product[]>("/products"));
    } catch (e) {
      setError(
        (e as Error).message + " バックエンドの起動を確認してください。",
      );
    } finally {
      setLoading(false);
    }
  }
  useEffect(() => {
    setManage(
      new URLSearchParams(window.location.search).get("view") === "manage",
    );
    void load();
  }, []);
  function add() {
    setEditProduct(null);
    setEditor(true);
  }
  const visible = products
    .filter(
      (p) =>
        (category === "すべて" || p.category === category) &&
        `${p.title} ${p.brand} ${p.note}`
          .toLowerCase()
          .includes(query.toLowerCase()),
    )
    .sort((a, b) => (sort === "old" ? a.id - b.id : b.id - a.id));
  return (
    <Shell onAdd={add} manage={manage}>
      <Layout
        height="auto"
        content={
          <LayoutContent padding={0}>
            <Center axis="horizontal">
              <VStack
                gap={8}
                maxWidth={1200}
                width="100%"
                paddingInline={6}
                paddingBlock={8}
              >
                <VStack gap={6} paddingBlock={8}>
                  <HStack hAlign="between" vAlign="center" wrap="wrap" gap={3}>
                    <Text type="supporting">A PERSONAL SELECT SHOP</Text>
                    <HStack gap={2} vAlign="center">
                      <Text type="supporting">EST. 2026</Text>
                      <Icon icon={ArrowUpRight} size="sm" />
                    </HStack>
                  </HStack>
                  <Heading
                    level={1}
                    style={{
                      fontSize: "clamp(2.8rem, 7vw, 5.6rem)",
                      lineHeight: 1.15,
                      letterSpacing: "-0.05em",
                      fontWeight: 500,
                    }}
                  >
                    好きなものに、
                    <br />
                    自分らしい視点を。
                  </Heading>
                  <HStack gap={6} wrap="wrap" hAlign="between" vAlign="end">
                    <VStack gap={3}>
                      <Text color="secondary">
                        暮らしに余白を、日常に小さなときめきを。
                        <br />
                        ドレの感性で集めた、とっておきのセレクト。
                      </Text>
                      <HStack gap={2} vAlign="center">
                        <Text type="label">ドレのセレクトショップ</Text>
                        <Text type="supporting">
                          / {products.length} selections
                        </Text>
                      </HStack>
                    </VStack>
                    <Button
                      label="あなただけのセレクトを追加"
                      variant="ghost"
                      icon={<Icon icon={Plus} />}
                      onClick={add}
                    />
                  </HStack>
                </VStack>
                <HStack
                  hAlign="between"
                  vAlign="center"
                  gap={4}
                  wrap="wrap"
                  style={{
                    borderTop: "solid var(--border-width) var(--color-border)",
                    paddingTop: "var(--spacing-6)",
                  }}
                >
                  <HStack gap={3} vAlign="center">
                    <Heading level={2}>
                      {manage ? "商品管理" : "The collection"}
                    </Heading>
                    <Text type="supporting">
                      {String(visible.length).padStart(2, "0")}
                    </Text>
                  </HStack>
                  <HStack gap={3} wrap="wrap">
                    <TextInput
                      label="商品を検索"
                      isLabelHidden
                      placeholder="セレクトを検索"
                      type="search"
                      startIcon={Search}
                      hasClear
                      value={query}
                      onChange={setQuery}
                    />
                    <Button
                      label={sort === "new" ? "新しい順" : "古い順"}
                      variant="ghost"
                      icon={<Icon icon={ArrowDown} />}
                      onClick={() =>
                        setSort((s) => (s === "new" ? "old" : "new"))
                      }
                    />
                  </HStack>
                </HStack>
                <TabList
                  value={category}
                  onChange={(v) => setCategory(String(v))}
                >
                  {[
                    ...new Set([
                      ...categories,
                      ...products.map((p) => p.category),
                    ]),
                  ].map((c) => (
                    <Tab key={c} value={c} label={c} />
                  ))}
                </TabList>
                {notice && (
                  <HStack role="status" hAlign="between">
                    <Text>{notice}</Text>
                    <Button
                      label="閉じる"
                      variant="ghost"
                      size="sm"
                      onClick={() => setNotice("")}
                    />
                  </HStack>
                )}
                {error ? (
                  <VStack gap={4} padding={6}>
                    <Text role="alert">{error}</Text>
                    <Button
                      label="もう一度読み込む"
                      onClick={() => void load()}
                    />
                  </VStack>
                ) : loading ? (
                  <Text role="status">セレクトを読み込んでいます…</Text>
                ) : visible.length === 0 ? (
                  <VStack gap={4} hAlign="center" paddingBlock={10}>
                    <Icon icon={Link2} size="lg" />
                    <Heading level={3}>
                      {products.length === 0
                        ? "最初の「好き」を、並べてみよう。"
                        : "条件に合う商品がありません"}
                    </Heading>
                    <Text color="secondary">
                      {products.length === 0
                        ? "お気に入りの商品URLを追加して、ショップをはじめましょう。"
                        : "検索キーワードやカテゴリを変更してみてください。"}
                    </Text>
                    <Button
                      label={
                        products.length === 0
                          ? "URLから商品を追加"
                          : "絞り込みをクリア"
                      }
                      onClick={() => {
                        if (products.length === 0) add();
                        else {
                          setQuery("");
                          setCategory("すべて");
                        }
                      }}
                    />
                  </VStack>
                ) : manage ? (
                  <VStack gap={0}>
                    {visible.map((p) => (
                      <HStack
                        key={p.id}
                        gap={4}
                        paddingBlock={4}
                        wrap="wrap"
                        vAlign="center"
                        style={{
                          borderBottom:
                            "solid var(--border-width) var(--color-border)",
                        }}
                      >
                        <VStack width="4rem">
                          <ProductImage
                            key={p.image_url}
                            src={p.image_url}
                            alt={p.title}
                            variant="thumbnail"
                          />
                        </VStack>
                        <StackItem size="fill">
                          <VStack gap={1}>
                            <Link href={`/products/${p.id}`} color="primary">
                              <Text weight="semibold">{p.title}</Text>
                            </Link>
                            <Text type="supporting">
                              {p.category} / {priceLabel(p)}
                            </Text>
                          </VStack>
                        </StackItem>
                        <Button
                          label="編集"
                          onClick={() => {
                            setEditProduct(p);
                            setEditor(true);
                          }}
                        />
                        <Button
                          label="詳細"
                          href={`/products/${p.id}`}
                          variant="ghost"
                        />
                      </HStack>
                    ))}
                  </VStack>
                ) : (
                  <Grid columns={{ minWidth: 260, max: 3 }} gap={6} rowGap={8}>
                    {visible.map((p) => (
                      <VStack key={p.id} gap={3}>
                        <NextLink
                          href={`/products/${p.id}`}
                          style={{
                            display: "block",
                            color: "inherit",
                            textDecoration: "none",
                            width: "100%",
                          }}
                        >
                          <VStack gap={3} width="100%">
                            <ProductImage
                              key={p.image_url}
                              src={p.image_url}
                              alt={p.title}
                            />
                            <HStack hAlign="between" gap={2} vAlign="center">
                              <Text type="supporting">
                                {p.brand || p.category}
                              </Text>
                              <Icon icon={ArrowUpRight} size="sm" />
                            </HStack>
                            <Heading level={3}>{p.title}</Heading>
                          </VStack>
                        </NextLink>
                        <Text type="supporting">{priceLabel(p)}</Text>
                        {p.note && (
                          <Text color="secondary" maxLines={2}>
                            {p.note}
                          </Text>
                        )}
                      </VStack>
                    ))}
                  </Grid>
                )}
                <VStack
                  hAlign="center"
                  paddingBlock={8}
                  gap={3}
                  style={{
                    borderTop: "solid var(--border-width) var(--color-border)",
                  }}
                >
                  <Text type="supporting">GOOD THINGS, THROUGH YOUR EYES.</Text>
                  <Text color="secondary">
                    誰かの定番より、あなたの「好き」を。
                  </Text>
                </VStack>
              </VStack>
            </Center>
          </LayoutContent>
        }
      />
      {editor && (
        <Editor
          product={editProduct}
          onClose={() => setEditor(false)}
          onSaved={(p) => {
            setEditor(false);
            setNotice(
              editProduct
                ? "セレクトを更新しました。"
                : "新しいセレクトを追加しました。",
            );
            setProducts((old) =>
              editProduct
                ? old.map((x) => (x.id === p.id ? p : x))
                : [p, ...old],
            );
            setCategory("すべて");
            setQuery("");
          }}
        />
      )}
    </Shell>
  );
}
