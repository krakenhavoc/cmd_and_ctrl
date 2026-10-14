// @vitest-environment jsdom
//
// ADR 0179 §7: "Cards from this game" lists the opponents' commanders
// and casts with Buy links and one list cart, names no one, carries the
// disclosure, and renders nothing without a shop or when hidden.

import { afterEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import GameCardsPanel from "./GameCardsPanel.svelte";
import { loadAppConfig, resetAppConfigForTests } from "../env";
import { updateSettings } from "../settings";
import { SHOP_DISCLOSURE } from "../shopLinks";
import { cleanup, flushSync, render } from "../test/render.svelte";

const realFetch = globalThis.fetch;

const BODY = {
  ready: true,
  opponents: [
    { commanders: ["Atraxa, Praetors' Voice"], cards: [{ name: "Sol Ring", casts: 2 }] },
    { commanders: ["Tymna the Weaver"], cards: [] },
  ],
};

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

async function setup(
  shops: unknown[],
  cards: unknown,
  status = 200,
): Promise<ReturnType<typeof vi.fn>> {
  resetAppConfigForTests();
  globalThis.fetch = vi.fn(
    async () => new Response(JSON.stringify({ env: "prod", shops })),
  ) as never;
  await loadAppConfig();
  flushSync();
  const f = vi.fn(
    async () =>
      new Response(JSON.stringify(status === 200 ? cards : { error: "nope" }), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
  );
  globalThis.fetch = f as never;
  return f;
}

const SHOPS = [{ id: "tcgplayer", template: "{url}&partner=c2" }];

afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  resetAppConfigForTests();
  updateSettings("display", "shopLinks", true);
});

describe("GameCardsPanel", () => {
  it("lists the opponents' cards with links, a cart and the disclosure, naming no one", async () => {
    await setup(SHOPS, BODY);
    const { container } = render(GameCardsPanel as never, { gameID: "g1" } as never);
    await settle();
    const text = container.textContent!.replace(/\s+/g, " ");
    expect(text).toContain("Cards from this game");
    expect(text).toContain("Opponent 1");
    expect(text).toContain("Opponent 2");
    expect(text).toContain("Atraxa, Praetors' Voice");
    expect(text).toContain("Sol Ring");
    expect(text).toContain("Buy these cards");
    expect(text).toContain(SHOP_DISCLOSURE);
    const hrefs = [...container.querySelectorAll("a")].map((a) => a.getAttribute("href") ?? "");
    expect(hrefs.some((h) => h.includes("massentry"))).toBe(true);
    expect(hrefs.some((h) => h.includes("q=Sol%20Ring"))).toBe(true);
    expect(container.querySelectorAll(".shop-note").length).toBe(1);
  });

  it("renders nothing without a shop, or when the viewer hid them", async () => {
    const f = await setup([], BODY);
    let r = render(GameCardsPanel as never, { gameID: "g1" } as never);
    await settle();
    expect(r.container.querySelector("[data-testid=game-cards]")).toBeNull();
    cleanup();

    await setup(SHOPS, BODY);
    updateSettings("display", "shopLinks", false);
    r = render(GameCardsPanel as never, { gameID: "g1" } as never);
    await settle();
    expect(r.container.querySelector("[data-testid=game-cards]")).toBeNull();
    expect(f).toBeDefined();
  });

  it("says so while the record is not ready, and shows a refusal", async () => {
    await setup(SHOPS, { ready: false, opponents: [] });
    let r = render(GameCardsPanel as never, { gameID: "g1" } as never);
    await settle();
    expect(r.container.textContent).toContain("not ready yet");
    expect(r.container.querySelector("a")).toBeNull();
    cleanup();

    await setup(SHOPS, {}, 409);
    r = render(GameCardsPanel as never, { gameID: "g1" } as never);
    await settle();
    expect(r.container.querySelector("[role=alert]")).not.toBeNull();
    expect(r.container.querySelector("a")).toBeNull();
  });
});
