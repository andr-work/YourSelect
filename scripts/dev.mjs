import { spawn, spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
const go =
  process.env.GO_BIN ||
  (spawnSync("go", ["version"]).status === 0
    ? "go"
    : existsSync("/tmp/yourselect-tools/go/bin/go")
      ? "/tmp/yourselect-tools/go/bin/go"
      : "go");
if (spawnSync(go, ["version"]).status !== 0) {
  console.error(
    "Go 1.25以降をインストールするか、GO_BINでGoのパスを指定してください。",
  );
  process.exit(1);
}
const env = {
  ...process.env,
  GOPATH: process.env.GOPATH || "/tmp/yourselect-tools/gopath",
  GOCACHE: process.env.GOCACHE || "/tmp/yourselect-tools/gocache",
};
const children = [
  spawn(go, ["run", "."], {
    cwd: new URL("../backend", import.meta.url),
    env,
    stdio: "inherit",
    detached: true,
  }),
  spawn(
    process.execPath,
    ["node_modules/next/dist/bin/next", "dev", "--hostname", "0.0.0.0"],
    { env, stdio: "inherit", detached: true },
  ),
];
let stopping = false;
function stop(code = 0) {
  if (stopping) return;
  stopping = true;
  for (const child of children) {
    try {
      process.kill(-child.pid, "SIGTERM");
    } catch {}
  }
  process.exitCode = code;
}
process.on("SIGINT", () => stop());
process.on("SIGTERM", () => stop());
for (const child of children) {
  child.on("error", (e) => {
    console.error(e.message);
    stop(1);
  });
  child.on("exit", (code) => {
    if (!stopping) stop(code || 0);
  });
}
