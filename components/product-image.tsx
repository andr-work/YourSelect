"use client";
import { useState } from "react";
import { AspectRatio } from "@astryxdesign/core/AspectRatio";
import { VStack } from "@astryxdesign/core/Layout";
import { Text } from "@astryxdesign/core/Text";
import { Icon } from "@astryxdesign/core/Icon";
import { ImageOff } from "lucide-react";

export function ProductImage({
  src,
  alt,
  variant = "collection",
}: {
  src: string;
  alt: string;
  variant?: "collection" | "thumbnail" | "detail";
}) {
  // Scope failures to the source so changing an image allows it to load again.
  const [failedSource, setFailedSource] = useState("");
  const square = variant === "thumbnail";
  const imageUrl = /^https?:\/\//i.test(src)
    ? `/api/images?${new URLSearchParams({ url: src, variant })}`
    : src;
  return (
    <VStack
      width="100%"
      style={{
        flexShrink: 0,
        alignSelf: "start",
        background: "var(--color-background-surface)",
      }}
    >
      <AspectRatio ratio={square ? 1 : 4 / 5} fit="contain">
        {src && failedSource !== imageUrl ? (
          <img
            src={imageUrl}
            alt={alt}
            width={square ? 160 : variant === "detail" ? 1200 : 800}
            height={square ? 160 : variant === "detail" ? 1500 : 1000}
            loading={variant === "detail" ? "eager" : "lazy"}
            decoding="async"
            onError={() => setFailedSource(imageUrl)}
          />
        ) : (
          <VStack gap={2} hAlign="center" vAlign="center" height="100%">
            <Icon icon={ImageOff} size="lg" />
            <Text type="supporting">画像はありません</Text>
          </VStack>
        )}
      </AspectRatio>
    </VStack>
  );
}
