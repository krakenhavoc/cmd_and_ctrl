// @vitest-environment jsdom
//
// #3558: the funding page, rendered with and without the operator's
// optional support links, plus the client's own check of those links.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Funding from "../routes/Funding.svelte";
import SiteFooter from "./components/SiteFooter.svelte";
import { loadAppConfig, parseConfig, resetAppConfigForTests } from "./env";
import { FUNDING_HASH, FUNDING_TITLE } from "./funding";
import { LEGAL_CONTACT_EMAIL, LEGAL_CONTACT_URL, OPERATOR_NAME } from "./legalNotice";
import { parseHash } from "./router";
import { SHOP_DISCLOSURE } from "./shopLinks";
import { GITHUB_SPONSORS_URL, parseSupport, validSupportUrl } from "./supportLinks";
import { cleanup, flushSync, render } from "./test/render.svelte";

const realFetch = globalThis.fetch;

function serve(body: unknown): void {
  globalThis.fetch = vi.fn(
    async () =>
      new Response(JSON.stringify(body), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
  ) as never;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 3; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => resetAppConfigForTests());
afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  resetAppConfigForTests();
  vi.restoreAllMocks();
});

function hrefs(container: HTMLElement): string[] {
  return [...container.querySelectorAll("a")].map((a) => a.getAttribute("href") ?? "");
}

describe("the funding page", () => {
  it("with no links configured: the promise, the costs, GitHub Sponsors, the operator", async () => {
    serve({ env: "prod", features: {}, shops: [], support: {} });
    await loadAppConfig();
    const { container } = render(Funding as never, {} as never);
    await settle();
    const text = container.textContent!.replace(/\s+/g, " ");

    expect(container.querySelector("h1")?.textContent).toBe(FUNDING_TITLE);
    expect(FUNDING_TITLE).toBe("How CMD & CTRL is funded");
    expect(text).toContain("free to play, and playing never requires paying");
    expect(text).toContain("What it costs to run");
    expect(text).toContain("It unlocks nothing in the game");
    expect(text).toContain(OPERATOR_NAME);
    expect(text).not.toMatch(/\$\s?\d/);

    const links = hrefs(container);
    expect(links).toContain(GITHUB_SPONSORS_URL);
    expect(links).toContain(LEGAL_CONTACT_URL);
    expect(text).toContain(LEGAL_CONTACT_EMAIL);
    expect(
      links.some((h) =>
        /(^|\.)(ko-fi|patreon)\.com$/.test(new URL(h, "https://x.invalid").hostname),
      ),
    ).toBe(false);
    // No shop configured: no shop sentence or commission note.
    expect(text).not.toContain(SHOP_DISCLOSURE);
  });

  it("with Ko-fi, Patreon and a shop configured: their links and the commission note", async () => {
    serve({
      env: "prod",
      features: {},
      shops: [{ id: "tcgplayer", name: "x", template: "{url}&partner=c2" }],
      support: { kofi: "https://ko-fi.com/cmdctrl", patreon: "https://www.patreon.com/cmdctrl" },
    });
    await loadAppConfig();
    const { container } = render(Funding as never, {} as never);
    await settle();
    const text = container.textContent!.replace(/\s+/g, " ");

    const anchors = [...container.querySelectorAll("a")];
    const kofi = anchors.find((a) => a.getAttribute("href") === "https://ko-fi.com/cmdctrl");
    const patreon = anchors.find(
      (a) => a.getAttribute("href") === "https://www.patreon.com/cmdctrl",
    );
    expect(kofi?.textContent).toBe("Ko-fi");
    expect(patreon?.textContent).toBe("Patreon");
    for (const a of [kofi, patreon]) {
      expect(a?.getAttribute("target")).toBe("_blank");
      expect(a?.getAttribute("rel")).toContain("noopener");
      expect(a?.getAttribute("rel")).toContain("noreferrer");
    }
    expect(text).toContain("third party");
    expect(text).toContain("TCGplayer");
    expect(text).toContain(SHOP_DISCLOSURE);
  });

  it("is linked from the footer and reachable by its hash", async () => {
    expect(parseHash(FUNDING_HASH)).toEqual({ name: "funding" });
    const { container } = render(SiteFooter as never, {} as never);
    await settle();
    const link = [...container.querySelectorAll("nav a")].find(
      (a) => a.getAttribute("href") === FUNDING_HASH,
    );
    expect(link).toBeTruthy();
  });

  it("holds nothing about plans or what is paid for", async () => {
    const { container } = render(Funding as never, {} as never);
    await settle();
    expect(container.textContent).not.toMatch(
      /subscri|premium|paid feature|\btier|pricing|monetiz/i,
    );
  });

  it("has no fixed width that would overflow a phone", () => {
    const { container } = render(Funding as never, {} as never);
    for (const el of container.querySelectorAll<HTMLElement>("[style]")) {
      expect(el.getAttribute("style")).not.toMatch(/width:\s*\d{3,}px/);
    }
  });
});

describe("support links from GET /config", () => {
  it("keeps only https URLs on the service's own domain", () => {
    expect(validSupportUrl("https://ko-fi.com/x", "ko-fi.com")).toBe(true);
    expect(validSupportUrl("https://www.ko-fi.com/x", "ko-fi.com")).toBe(true);
    expect(validSupportUrl("http://ko-fi.com/x", "ko-fi.com")).toBe(false);
    expect(validSupportUrl("https://ko-fi.com.evil.example/x", "ko-fi.com")).toBe(false);
    expect(validSupportUrl("https://ko-fi.com@evil.example/x", "ko-fi.com")).toBe(false);
    expect(validSupportUrl("https://user:pw@ko-fi.com/x", "ko-fi.com")).toBe(false);
    expect(validSupportUrl("https://ko-fi.com:444/x", "ko-fi.com")).toBe(false);
    expect(validSupportUrl("javascript:alert(1)", "ko-fi.com")).toBe(false);
    expect(validSupportUrl("https://patreon.com/x", "ko-fi.com")).toBe(false);

    expect(
      parseSupport({
        kofi: "https://patreon.com/swapped",
        patreon: "https://patreon.com/ok",
        other: "https://ko-fi.com/x",
      }),
    ).toEqual({ patreon: "https://patreon.com/ok" });
    expect(parseSupport("nope")).toEqual({});
    expect(parseSupport(null)).toEqual({});
  });

  it("rides on parseConfig for production and dev", () => {
    const support = { kofi: "https://ko-fi.com/c2" };
    expect(parseConfig({ env: "prod", support }).support).toEqual(support);
    expect(parseConfig({ env: "dev", support }).support).toEqual(support);
    expect(parseConfig({ env: "prod" }).support).toEqual({});
  });
});
