// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import { applyFavicon, faviconDataUrl, faviconSvg, usableColour } from "./favicon";

const c = { bg: "#0b0a09", accent: "#d9b45c", mark: "#f1ece2" };

describe("favicon", () => {
  it("paints the tile, the Cs and the diamond", () => {
    const svg = faviconSvg(c);
    expect(svg).toContain('rx="12" fill="#0b0a09"');
    expect(svg).toContain('fill="#d9b45c"');
    expect(svg).toContain('d="M0-38L38 0L0 38L-38 0Z" fill="#f1ece2"');
  });

  it("keeps the geometry of public/icons/favicon.svg", () => {
    const file = readFileSync("public/icons/favicon.svg", "utf8");
    const paths = (s: string) => [...s.matchAll(/ d="([^"]+)"/g)].map((m) => m[1]);
    expect(paths(faviconSvg(c))).toEqual(paths(file));
    const transforms = (s: string) => [...s.matchAll(/transform="([^"]+)"/g)].map((m) => m[1]);
    expect(transforms(faviconSvg(c))).toEqual(transforms(file));
  });

  it("is a data URL that decodes to the svg", () => {
    const url = faviconDataUrl(c);
    expect(url.startsWith("data:image/svg+xml,")).toBe(true);
    expect(decodeURIComponent(url.slice("data:image/svg+xml,".length))).toBe(faviconSvg(c));
  });

  it("rejects colours that will not resolve in a standalone image", () => {
    expect(usableColour("")).toBe(false);
    expect(usableColour("color-mix(in srgb, red, blue)")).toBe(false);
    expect(usableColour("rgb(1, 2, 3)")).toBe(true);
    expect(usableColour("#fff")).toBe(true);
  });

  it("sets link[rel=icon] and skips on an empty colour", () => {
    document.head.innerHTML = '<link rel="icon" href="/icons/favicon.svg">';
    applyFavicon(document, { ...c, accent: "" });
    expect(document.querySelector('link[rel="icon"]')?.getAttribute("href")).toBe(
      "/icons/favicon.svg",
    );
    applyFavicon(document, c);
    expect(document.querySelector('link[rel="icon"]')?.getAttribute("href")).toBe(
      faviconDataUrl(c),
    );
  });
});
