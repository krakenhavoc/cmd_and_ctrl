// The type is self-hosted (#3345): no third-party font origin in the shell,
// every @font-face file exists with its licence, and the service worker's
// plugin precaches it.
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const root = resolve(__dirname, "../..");
const read = (p: string) => readFileSync(resolve(root, p), "utf8");

describe("self-hosted fonts", () => {
  it("index.html names no external font origin", () => {
    const html = read("index.html");
    expect(html).not.toMatch(/fonts\.googleapis\.com|fonts\.gstatic\.com/);
  });

  it("every @font-face file exists, is precached, and each family has a licence", () => {
    const css = read("src/app.css");
    const files = [...css.matchAll(/url\("(\/fonts\/[^"]+\.woff2)"\)/g)].map((m) => m[1]);
    expect(files.length).toBe(3);
    const plugin = read("vite-plugin-sw.ts");
    for (const f of files) {
      expect(existsSync(resolve(root, "public" + f)), f).toBe(true);
      expect(plugin, f).toContain(`"${f}"`);
    }
    for (const lic of ["bricolage-grotesque", "instrument-sans", "jetbrains-mono"]) {
      expect(existsSync(resolve(root, `public/fonts/OFL-${lic}.txt`)), lic).toBe(true);
    }
  });
});
