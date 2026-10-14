// @vitest-environment jsdom
//
// ADR 0179: a group of shop links renders only with a configured shop
// and the viewer's setting on, carries the disclosure, opens in a new
// tab without a referrer, and hides from its own button.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

import ShopLinks from "./components/ShopLinks.svelte";
import { loadAppConfig, resetAppConfigForTests } from "./env";
import { settings, updateSettings } from "./settings";
import { SHOP_DISCLOSURE } from "./shopLinks";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

async function configure(shops: unknown[]): Promise<void> {
  resetAppConfigForTests();
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue({ ok: true, json: async () => ({ env: "prod", shops }) }),
  );
  await loadAppConfig();
  flushSync();
}

const SHOPS = [
  { id: "tcgplayer", template: "https://partner.example/c/1?u={url}" },
  { id: "cardkingdom", template: "{url}&partner=c2" },
];

beforeEach(() => {
  updateSettings("display", "shopLinks", true);
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  resetAppConfigForTests();
  updateSettings("display", "shopLinks", true);
});

describe("ShopLinks", () => {
  it("renders nothing when no shop is configured", async () => {
    await configure([]);
    const r = render(ShopLinks, { card: "Sol Ring" });
    expect(r.container.querySelector("a")).toBeNull();
    expect(r.container.textContent).not.toContain(SHOP_DISCLOSURE);
  });

  it("renders one link per shop, with the disclosure", async () => {
    await configure(SHOPS);
    const r = render(ShopLinks, { list: [{ name: "Sol Ring", count: 1 }], label: "Buy this deck" });
    const links = [...r.container.querySelectorAll("a")];
    expect(links.map((a) => a.textContent?.trim())).toEqual(["TCGplayer", "Card Kingdom"]);
    for (const a of links) {
      expect(a.getAttribute("target")).toBe("_blank");
      expect(a.getAttribute("rel")).toContain("noreferrer");
      expect(a.getAttribute("rel")).toContain("sponsored");
      expect(a.href.startsWith("https://")).toBe(true);
    }
    expect(links[1].href).toBe("https://www.cardkingdom.com/builder?c=1%20Sol%20Ring&partner=c2");
    expect(r.container.textContent).toContain(SHOP_DISCLOSURE);
    expect(r.container.textContent).toContain("Buy this deck");
  });

  it("leaves the disclosure to the caller with note={false}", async () => {
    await configure(SHOPS);
    const r = render(ShopLinks, { card: "Sol Ring", compact: true, note: false });
    expect(r.container.querySelectorAll("a")).toHaveLength(2);
    expect(r.container.textContent).not.toContain(SHOP_DISCLOSURE);
  });

  it("hides every link from its own button, and stays hidden", async () => {
    await configure(SHOPS);
    const r = render(ShopLinks, { card: "Sol Ring" });
    const hide = [...r.container.querySelectorAll("button")].find((b) =>
      b.textContent?.includes("Hide shop links"),
    );
    expect(hide).toBeDefined();
    click(hide!);
    expect(get(settings).display.shopLinks).toBe(false);
    expect(r.container.querySelector("a")).toBeNull();
    const again = render(ShopLinks, { card: "Lightning Bolt" });
    expect(again.container.querySelector("a")).toBeNull();
  });
});
