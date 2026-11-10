"use client";
// Adapted from Astryx detail-page: summary header, two-column detail and action rail.
import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  VStack,
  HStack,
  Layout,
  LayoutContent,
} from "@astryxdesign/core/Layout";
import { Center } from "@astryxdesign/core/Center";
import { Grid } from "@astryxdesign/core/Grid";
import { Text, Heading } from "@astryxdesign/core/Text";
import { Button } from "@astryxdesign/core/Button";
import { Dialog, DialogHeader } from "@astryxdesign/core/Dialog";
import { Icon } from "@astryxdesign/core/Icon";
import { ArrowLeft, ArrowUpRight, Pencil, Trash2 } from "lucide-react";
import { Shell } from "../../../components/shell";
import { Editor } from "../../../components/editor";
import { ProductImage } from "../../../components/product-image";
import { api, Product, priceLabel, sourceLabel } from "../../../lib/api";
export default function Detail({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const router = useRouter();
  const [product, setProduct] = useState<Product | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    setLoading(true);
    api<Product>(`/products/${id}`)
      .then(setProduct)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);
  async function remove() {
    setBusy(true);
    setError("");
    try {
      const res = await fetch(`/api/products/${id}`, { method: "DELETE" });
      if (!res.ok) {
        const body = await res.json();
        throw new Error(body.error);
      }
      router.push("/?view=manage");
    } catch (e) {
      setError((e as Error).message);
      setBusy(false);
    }
  }
  return (
    <Shell>
      <Layout
        height="auto"
        content={
          <LayoutContent padding={0}>
            <Center axis="horizontal">
              <VStack
                maxWidth={1200}
                width="100%"
                gap={8}
                paddingInline={6}
                paddingBlock={8}
              >
                <HStack>
                  <Button
                    label="コレクションに戻る"
                    href="/"
                    variant="ghost"
                    icon={<Icon icon={ArrowLeft} />}
                  />
                </HStack>
                {error && <Text role="alert">{error}</Text>}
                {loading ? (
                  <Text role="status">商品を読み込んでいます…</Text>
                ) : product ? (
                  <>
                    <Grid columns={{ minWidth: 320, max: 2 }} gap={8}>
                      <ProductImage
                        key={product.image_url}
                        src={product.image_url}
                        alt={product.title}
                        variant="detail"
                      />
                      <VStack gap={6} paddingBlock={4}>
                        <Text type="supporting">
                          {product.category} /{" "}
                          {product.brand || sourceLabel(product.source_url)}
                        </Text>
                        <Heading level={1}>{product.title}</Heading>
                        <Text type="large">{priceLabel(product)}</Text>
                        <Text color="secondary">
                          {product.description ||
                            "商品についての詳しい情報は、販売サイトをご覧ください。"}
                        </Text>
                        <VStack
                          gap={3}
                          padding={6}
                          style={{
                            background: "var(--color-background-muted)",
                          }}
                        >
                          <Text type="supporting">WHY I SELECTED</Text>
                          <Text>
                            {product.note ||
                              "まだセレクトの理由は書かれていません。"}
                          </Text>
                          <Text type="supporting">— ドレ</Text>
                        </VStack>
                        <Button
                          label="販売サイトで見る"
                          href={product.source_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          variant="primary"
                          size="lg"
                          icon={<Icon icon={ArrowUpRight} />}
                        />
                        <Text type="supporting">
                          {sourceLabel(product.source_url)} ·
                          価格・在庫は販売サイトでご確認ください。
                        </Text>
                        <HStack gap={3} wrap="wrap">
                          <Button
                            label="商品を編集"
                            icon={<Icon icon={Pencil} />}
                            onClick={() => setEditing(true)}
                          />
                          <Button
                            label="削除"
                            variant="ghost"
                            icon={<Icon icon={Trash2} />}
                            onClick={() => setDeleting(true)}
                          />
                        </HStack>
                        <Text type="supporting">
                          追加日{" "}
                          {new Date(product.created_at).toLocaleDateString(
                            "ja-JP",
                          )}
                        </Text>
                      </VStack>
                    </Grid>
                  </>
                ) : null}
              </VStack>
            </Center>
          </LayoutContent>
        }
      />
      {editing && product && (
        <Editor
          product={product}
          onClose={() => setEditing(false)}
          onSaved={(p) => {
            setProduct(p);
            setEditing(false);
          }}
        />
      )}
      {deleting && (
        <Dialog
          isOpen
          onOpenChange={(v) => {
            if (!v && !busy) setDeleting(false);
          }}
          purpose="form"
        >
          <DialogHeader title="このセレクトを削除しますか？" />
          <VStack gap={5} padding={6}>
            <Text>
              「{product?.title}
              」をショップから削除します。この操作は取り消せません。
            </Text>
            {error && <Text role="alert">{error}</Text>}
            <HStack gap={3} hAlign="end">
              <Button
                label="キャンセル"
                onClick={() => setDeleting(false)}
                isDisabled={busy}
              />
              <Button
                label="削除する"
                variant="destructive"
                isLoading={busy}
                onClick={() => void remove()}
              />
            </HStack>
          </VStack>
        </Dialog>
      )}
    </Shell>
  );
}
