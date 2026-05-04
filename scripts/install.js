#!/usr/bin/env node

const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const packageRoot = path.join(__dirname, "..");
const vendorDir = path.join(packageRoot, "vendor");
const binaryName = process.platform === "win32" ? "talizen.exe" : "talizen";
const targetBinary = path.join(vendorDir, binaryName);

const owner = process.env.TALIZEN_CLI_GITHUB_OWNER || "talizen";
const repo = process.env.TALIZEN_CLI_GITHUB_REPO || "talizen-cli";
const version = (process.env.npm_package_version || readPackageVersion()).trim();
const tag = version.startsWith("v") ? version : `v${version}`;
const checkOnly = process.argv.includes("--check");

main().catch((error) => {
  console.error(`Failed to install Talizen CLI: ${error.message}`);
  process.exit(1);
});

async function main() {
  const platform = getPlatform(process.platform);
  const arch = getArch(process.arch);
  const ext = platform === "windows" ? "zip" : "tar.gz";
  const archiveName = `talizen_${version}_${platform}_${arch}.${ext}`;

  if (checkOnly) {
    console.log(`Would install ${archiveName} from ${owner}/${repo}@${tag}`);
    return;
  }

  fs.mkdirSync(vendorDir, { recursive: true });

  if (process.env.TALIZEN_CLI_SKIP_DOWNLOAD === "1") {
    console.log("Skipping Talizen CLI binary download.");
    return;
  }

  const archiveURL = releaseURL(archiveName);
  const checksumURL = releaseURL("checksums.txt");
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "talizen-cli-"));
  const archivePath = path.join(tmpDir, archiveName);

  try {
    await download(archiveURL, archivePath);
    await verifyChecksum(checksumURL, archivePath, archiveName);
    extractArchive(archivePath, tmpDir, platform);
    const extractedBinary = path.join(tmpDir, binaryName);

    if (!fs.existsSync(extractedBinary)) {
      throw new Error(`release archive did not contain ${binaryName}`);
    }

    fs.copyFileSync(extractedBinary, targetBinary);
    if (process.platform !== "win32") {
      fs.chmodSync(targetBinary, 0o755);
    }
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

function readPackageVersion() {
  const pkg = JSON.parse(fs.readFileSync(path.join(packageRoot, "package.json"), "utf8"));
  return pkg.version;
}

function getPlatform(value) {
  switch (value) {
    case "darwin":
      return "darwin";
    case "linux":
      return "linux";
    case "win32":
      return "windows";
    default:
      throw new Error(`unsupported platform: ${value}`);
  }
}

function getArch(value) {
  switch (value) {
    case "x64":
      return "amd64";
    case "arm64":
      return "arm64";
    default:
      throw new Error(`unsupported architecture: ${value}`);
  }
}

function releaseURL(name) {
  return `https://github.com/${owner}/${repo}/releases/download/${tag}/${name}`;
}

async function download(url, destination) {
  const response = await fetch(url, {
    headers: {
      "user-agent": "@talizen/talizen-cli installer",
    },
    redirect: "follow",
  });

  if (!response.ok) {
    throw new Error(`download failed (${response.status}) for ${url}`);
  }

  const data = Buffer.from(await response.arrayBuffer());
  fs.writeFileSync(destination, data);
}

async function verifyChecksum(checksumURL, archivePath, archiveName) {
  const response = await fetch(checksumURL, {
    headers: {
      "user-agent": "@talizen/talizen-cli installer",
    },
    redirect: "follow",
  });

  if (!response.ok) {
    throw new Error(`checksum download failed (${response.status}) for ${checksumURL}`);
  }

  const checksums = await response.text();
  const expected = findChecksum(checksums, archiveName);
  const actual = crypto.createHash("sha256").update(fs.readFileSync(archivePath)).digest("hex");

  if (actual !== expected) {
    throw new Error(`checksum mismatch for ${archiveName}`);
  }
}

function findChecksum(checksums, archiveName) {
  for (const line of checksums.split(/\r?\n/)) {
    const parts = line.trim().split(/\s+/);
    if (parts.length >= 2 && parts[1] === archiveName) {
      return parts[0];
    }
  }
  throw new Error(`checksum not found for ${archiveName}`);
}

function extractArchive(archivePath, tmpDir, platform) {
  const command = platform === "windows" ? "powershell.exe" : "tar";
  const args =
    platform === "windows"
      ? [
          "-NoProfile",
          "-Command",
          `Expand-Archive -LiteralPath ${powershellQuote(archivePath)} -DestinationPath ${powershellQuote(tmpDir)} -Force`,
        ]
      : ["-xzf", archivePath, "-C", tmpDir];

  const result = spawnSync(command, args, { stdio: "inherit" });
  if (result.error) {
    throw result.error;
  }
  if (result.status !== 0) {
    throw new Error(`${command} exited with status ${result.status}`);
  }
}

function powershellQuote(value) {
  return `'${value.replace(/'/g, "''")}'`;
}
