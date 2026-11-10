"use client";
import { VStack, HStack, StackItem } from "@astryxdesign/core/Layout";
import { Link } from "@astryxdesign/core/Link";
import { Text } from "@astryxdesign/core/Text";
import { Button } from "@astryxdesign/core/Button";
import { Icon } from "@astryxdesign/core/Icon";
import { ArrowUpRight, Plus, Store, LayoutGrid } from "lucide-react";
export function Shell({
  children,
  onAdd,
  manage = false,
}: {
  children: React.ReactNode;
  onAdd?: () => void;
  manage?: boolean;
}) {
  return (
    <VStack gap={0} minHeight="100dvh">
      <HStack
        as="header"
        paddingInline={6}
        paddingBlock={4}
        gap={5}
        vAlign="center"
        wrap="wrap"
        style={{
          borderBottom: "solid var(--border-width) var(--color-border)",
          position: "sticky",
          top: 0,
          zIndex: 10,
          background: "var(--color-background-surface)",
        }}
      >
        <Link href="/" color="primary" size="2xl" weight="bold">
          yourselect<Text color="accent">.</Text>
        </Link>
        <StackItem size="fill">
          <HStack gap={1}>
            <Button
              href="/"
              label="マイショップ"
              variant={manage ? "ghost" : "secondary"}
              icon={<Icon icon={Store} />}
            />
            <Button
              href="/?view=manage"
              label="商品管理"
              variant={manage ? "secondary" : "ghost"}
              icon={<Icon icon={LayoutGrid} />}
            />
          </HStack>
        </StackItem>
        {onAdd && (
          <Button
            label="商品を追加"
            variant="primary"
            size="sm"
            icon={<Icon icon={Plus} />}
            onClick={onAdd}
          />
        )}
        <HStack gap={2} vAlign="center">
          <VStack
            hAlign="center"
            style={{
              background: "var(--color-background-muted)",
              borderRadius: "50%",
              width: "2.3rem",
              height: "2.3rem",
              justifyContent: "center",
            }}
          >
            <Text weight="semibold">ド</Text>
          </VStack>
          <Text type="label">ドレ</Text>
        </HStack>
      </HStack>
      {children}
      <HStack
        as="footer"
        padding={6}
        gap={4}
        wrap="wrap"
        hAlign="between"
        style={{
          marginTop: "auto",
          borderTop: "solid var(--border-width) var(--color-border)",
        }}
      >
        <Text type="supporting">yourselect.　好きが、ショップになる。</Text>
        <Text type="supporting">
          CURATED BY ドレ <Icon icon={ArrowUpRight} size="sm" />
        </Text>
      </HStack>
    </VStack>
  );
}
