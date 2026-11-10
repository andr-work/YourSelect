"use client";
import { useState } from "react";
import { VStack, HStack } from "@astryxdesign/core/Layout";
import { Dialog, DialogHeader } from "@astryxdesign/core/Dialog";
import { Text } from "@astryxdesign/core/Text";
import { TextInput } from "@astryxdesign/core/TextInput";
import { Button } from "@astryxdesign/core/Button";
import { Icon } from "@astryxdesign/core/Icon";
import { Link2, ArrowRight } from "lucide-react";
import { api, blankProduct, categories, Product } from "../lib/api";
export function Editor({
  product,
  onClose,
  onSaved,
}: {
  product: Product | null;
  onClose: () => void;
  onSaved: (p: Product) => void;
}) {
  const [draft, setDraft] = useState<Product>(product || { ...blankProduct });
  const [step, setStep] = useState(product ? "edit" : "url");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const update = (key: keyof Product, value: string) =>
    setDraft((p) => ({ ...p, [key]: value }));
  async function preview() {
    setBusy(true);
    setError("");
    try {
      const p = await api<Product>("/preview", {
        method: "POST",
        body: JSON.stringify({ url: draft.source_url }),
      });
      setDraft({ ...blankProduct, ...p });
      setStep("edit");
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  async function save() {
    setBusy(true);
    setError("");
    try {
      const p = await api<Product>(
        product ? `/products/${product.id}` : "/products",
        { method: product ? "PUT" : "POST", body: JSON.stringify(draft) },
      );
      onSaved(p);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <Dialog
      isOpen
      onOpenChange={(v) => {
        if (!v && !busy) onClose();
      }}
      purpose="form"
      width="min(92vw, 38rem)"
    >
      <DialogHeader
        title={product ? "セレクトを編集" : "新しいセレクトを追加"}
        subtitle={
          step === "url"
            ? "見つけた「好き」を、URLひとつで。"
            : "商品の情報と、あなたの視点を添えて。"
        }
        onOpenChange={(v) => {
          if (!v && !busy) onClose();
        }}
      />
      <VStack gap={5} padding={6}>
        <TextInput
          label="商品ページのURL"
          type="url"
          placeholder="https://shop.example.com/products/..."
          value={draft.source_url}
          onChange={(v) => update("source_url", v)}
          isDisabled={busy}
          onEnter={() => {
            if (step === "url") void preview();
          }}
        />
        {step === "url" ? (
          <>
            <Text color="secondary" type="supporting">
              商品名・画像・価格をリンク先から取得します。保存前に内容を編集できます。
            </Text>
            <Button
              label="URLから商品情報を取得"
              variant="primary"
              icon={<Icon icon={Link2} />}
              isLoading={busy}
              onClick={() => void preview()}
            />
            <Button
              label="手入力で登録する"
              variant="ghost"
              isDisabled={busy}
              onClick={() => {
                setError("");
                setStep("edit");
              }}
            />
          </>
        ) : (
          <>
            <TextInput
              label="商品名"
              value={draft.title}
              onChange={(v) => update("title", v)}
              isRequired
            />
            <HStack gap={3}>
              <TextInput
                label="ブランド / ショップ"
                value={draft.brand}
                onChange={(v) => update("brand", v)}
                width="100%"
              />
              <TextInput
                label="カテゴリ"
                value={draft.category}
                onChange={(v) => update("category", v)}
                width="100%"
              />
            </HStack>
            <Text type="supporting">
              カテゴリ例：{categories.slice(1).join(" / ")}
            </Text>
            <HStack gap={3}>
              <TextInput
                label="価格"
                inputMode="decimal"
                value={draft.price}
                onChange={(v) => update("price", v)}
                width="100%"
              />
              <TextInput
                label="通貨"
                value={draft.currency}
                onChange={(v) => update("currency", v)}
                width="40%"
              />
            </HStack>
            <TextInput
              label="画像URL"
              type="url"
              value={draft.image_url}
              onChange={(v) => update("image_url", v)}
            />
            <TextInput
              label="商品説明"
              value={draft.description}
              onChange={(v) => update("description", v)}
            />
            <TextInput
              label="セレクトした理由"
              placeholder="どんなところに惹かれましたか？"
              value={draft.note}
              onChange={(v) => update("note", v)}
            />
            <HStack gap={3} hAlign="end">
              <Button label="キャンセル" onClick={onClose} isDisabled={busy} />
              <Button
                label={product ? "変更を保存" : "ショップに追加"}
                variant="primary"
                icon={<Icon icon={ArrowRight} />}
                isLoading={busy}
                onClick={() => void save()}
              />
            </HStack>
          </>
        )}
        {error && (
          <Text as="p" role="alert">
            {error}
          </Text>
        )}
      </VStack>
    </Dialog>
  );
}
