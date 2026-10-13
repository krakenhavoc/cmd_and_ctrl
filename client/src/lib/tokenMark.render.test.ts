// @vitest-environment jsdom
//
// #3043: a token copy wears the copied card's art and frame, so the
// board needs a mark that says which permanent is the token. Every
// token carries one; a copy of a real card says "TOKEN COPY".

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { hoveredCard } from "./cardTypes";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const base = {
  owner: "me",
  controller: "me",
  known_by_you: true,
  power: 6,
  toughness: 6,
};

const realTyrant = (): CardView =>
  ({
    ...base,
    instance_id: "real",
    name: "Vaultborn Tyrant",
    scryfall_id: "aaaa-bbbb",
    type_line: "Creature — Dinosaur",
  }) as unknown as CardView;

const copyTyrant = (): CardView =>
  ({
    ...base,
    instance_id: "copy",
    name: "Vaultborn Tyrant",
    scryfall_id: "aaaa-bbbb",
    type_line: "Token Creature — Dinosaur",
    is_token: true,
    token_copy: true,
  }) as unknown as CardView;

const plainToken = (): CardView =>
  ({
    ...base,
    instance_id: "tok",
    name: "Soldier",
    type_line: "Token Creature — Soldier",
    power: 1,
    toughness: 1,
    is_token: true,
  }) as unknown as CardView;

function mount(card: CardView, size = "small") {
  return render(Card as never, { card, size } as Record<string, unknown>);
}

afterEach(cleanup);

describe("the token mark on a board tile", () => {
  it("says TOKEN COPY on a token copy of a real card", () => {
    const { container } = mount(copyTyrant());
    const badge = container.querySelector(".badge.token-mark");
    expect(badge?.textContent?.trim()).toBe("TOKEN COPY");
    expect(badge?.classList.contains("copy")).toBe(true);
  });

  it("says TOKEN on a plain token, art or no art", () => {
    const { container } = mount(plainToken());
    const badge = container.querySelector(".badge.token-mark");
    expect(badge?.textContent?.trim()).toBe("TOKEN");
    expect(badge?.classList.contains("copy")).toBe(false);
  });

  it("says TOKEN on a plain token that has a token printing", () => {
    const { container } = mount({ ...plainToken(), scryfall_id: "tok-print" } as CardView);
    expect(container.querySelector(".badge.token-mark")?.textContent?.trim()).toBe("TOKEN");
  });

  it("leaves the real card alone", () => {
    const { container } = mount(realTyrant());
    expect(container.querySelector(".badge.token-mark")).toBeNull();
  });

  it("falls back to the type line for a view that predates is_token", () => {
    const old = { ...plainToken(), is_token: undefined } as CardView;
    const { container } = mount(old);
    expect(container.querySelector(".badge.token-mark")).not.toBeNull();
  });

  it("is present on a compact art tile", () => {
    const { container } = mount(copyTyrant(), "tiny");
    expect(container.querySelector(".badge.token-mark")?.textContent?.trim()).toBe("TOKEN COPY");
  });

  it("clears the PHASED badge when both apply", () => {
    const { container } = mount({ ...copyTyrant(), phased_out: true } as CardView);
    expect(container.querySelector(".badge.token-mark.with-phased")).not.toBeNull();
    expect(container.querySelector(".badge.phased")).not.toBeNull();
  });
});

const view: GameView = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
} as unknown as GameView;

describe("the token chip in the hover panel", () => {
  let realFetch: unknown;
  beforeEach(() => {
    hoveredCard.set(null);
    const g = globalThis as Record<string, unknown>;
    realFetch = g.fetch;
    g.fetch = () =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ name: "Bear", type_line: "Creature", oracle_text: "" }),
      } as unknown as Response);
  });
  afterEach(() => {
    (globalThis as Record<string, unknown>).fetch = realFetch;
  });

  async function chip(card: CardView): Promise<string | null> {
    const { container } = render(HoverZoomOverlay as never, { view } as never);
    hoveredCard.set(card);
    for (let i = 0; i < 6; i++) {
      await Promise.resolve();
      flushSync();
    }
    return container.querySelector(".state-token")?.textContent?.trim() ?? null;
  }

  it("says token copy for a copy", async () => {
    expect(await chip(copyTyrant())).toBe("token copy");
  });
  it("says token for a plain token", async () => {
    expect(await chip(plainToken())).toBe("token");
  });
  it("says nothing for a real card", async () => {
    expect(await chip(realTyrant())).toBeNull();
  });
});
