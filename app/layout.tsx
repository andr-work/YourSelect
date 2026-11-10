import type { Metadata } from "next";
import "@astryxdesign/core/reset.css";
import "@astryxdesign/core/astryx.css";
import { Providers } from "../components/providers";
export const metadata: Metadata = {
  title: "yourselect — 好きが、ショップになる。",
  description: "ドレの視点で選ぶ、日常を少し豊かにするもの。",
};
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ja">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
