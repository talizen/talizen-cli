#!/usr/bin/env node

const { spawnSync } = require("node:child_process");
const path = require("node:path");

const exe = process.platform === "win32" ? "talizen.exe" : "talizen";
const binary = path.join(__dirname, "..", "vendor", exe);

const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
});

if (result.error) {
  if (result.error.code === "ENOENT") {
    console.error(
      "Talizen CLI binary is missing. Reinstall talizen-cli and try again.",
    );
  } else {
    console.error(result.error.message);
  }
  process.exit(1);
}

if (typeof result.status === "number") {
  process.exit(result.status);
}

process.exit(result.signal ? 1 : 0);
