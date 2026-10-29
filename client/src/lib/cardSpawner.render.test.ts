// @vitest-environment jsdom
//
// The spawner's two modes (ADR 0075 §2.4).
//
// The component moved out of components/dev/ in S35 and grew a
// production half, so what is worth pinning is the seam between them:
// the dev tool must keep behaving exactly as it did — the dev route
// cannot make a token and must not be offered a Tokens tab — while
// the managed one gains tokens and the zone rule that comes with
// them.

import { describe, it, expect, afterEach, vi } from "vitest";

import CardSpawner from "./components/CardSpawner.svelte";
import type { GameView } from "./protocol";
import { render, cleanup, click, flushSync } from "./test/render.svelte";

// The component reaches for the network on mount (the token list) and
// on every keystroke (the card search). Neither is what these tests
// are about.
const searchDevCardsMock = vi.hoisted(() => vi.fn(async (): Promise<unknown[]> => []));

vi.mock("./api", () => ({
  searchDevCards: searchDevCardsMock,
  searchSpawnCards: vi.fn(async () => []),
  spawnDevCard: vi.fn(async () => ({ spawned: [], name: "", zone: "", count: 0 })),
  spawnOnTable: vi.fn(async () => ({ spawned: [], name: "", zone: "", count: 0 })),
  fetchSpawnTokens: vi.fn(async () => ["Treasure", "1/1 white Soldier"]),
}));

afterEach(cleanup);

const snapshot = {
  seats: [
    { id: "p1", name: "Alice" },
    { id: "p2", name: "Bob" },
  ],
} as unknown as GameView;

function mount(managed: boolean) {
  return render(CardSpawner as never, { gameID: "g1", snapshot, managed } as never);
}

function tabs(container: HTMLElement): HTMLButtonElement[] {
  return [...container.querySelectorAll<HTMLButtonElement>('[role="tab"]')];
}

function zoneOptions(container: HTMLElement): string[] {
  const select = container.querySelectorAll("select")[1];
  return [...select.querySelectorAll("option")].map((o) => o.value);
}

describe("the spawner in dev mode", () => {
  it("offers no Tokens tab — the dev route can only resolve a printing", () => {
    const { container } = mount(false);
    expect(tabs(container)).toHaveLength(0);
  });

  it("keeps every zone the dev spawner has always had", () => {
    const { container } = mount(false);
    expect(zoneOptions(container)).toEqual([
      "battlefield",
      "hand",
      "graveyard",
      "exile",
      "library",
      "command",
    ]);
  });
});

describe("the spawner in managed mode", () => {
  it("offers Cards and Tokens, starting on Cards", () => {
    const { container } = mount(true);
    expect(tabs(container).map((t) => t.textContent?.trim())).toEqual(["Cards", "Tokens"]);
    expect(tabs(container)[0].getAttribute("aria-selected")).toBe("true");
  });

  it("lists the server's token templates once the tab is opened", async () => {
    const { container } = mount(true);
    click(tabs(container)[1]);
    // The fetch is lazy; let its promise settle and re-render.
    await Promise.resolve();
    await Promise.resolve();
    flushSync();
    const names = [...container.querySelectorAll(".result.token")].map((b) =>
      b.textContent?.trim(),
    );
    expect(names).toContain("Treasure");
  });

  it("locks a token to the battlefield", () => {
    // CR 704.5d — a token in any other zone ceases to exist at the
    // next state-based action check, so the option would appear to
    // work and then silently undo itself.
    const { container } = mount(true);
    click(tabs(container)[1]);
    expect(zoneOptions(container)).toEqual(["battlefield"]);
  });

  it("drops the Commander tickbox on the tokens tab — a token is never a commander", () => {
    const { container } = mount(true);
    expect(container.querySelector('input[type="checkbox"]')).not.toBeNull();
    click(tabs(container)[1]);
    expect(container.querySelector('input[type="checkbox"]')).toBeNull();
  });

  it("says that a spawn is public and free to take back", () => {
    const { container } = mount(true);
    expect(container.textContent).toMatch(/named in the game log/i);
  });

  it("gives the dev tool no such promise — the dev route logs nothing", () => {
    const { container } = mount(false);
    expect(container.textContent).not.toMatch(/named in the game log/i);
  });
});

// #4424: a double-faced card matched on more than one of its names used to
// come back from the search more than once. The result list is keyed by
// card id, so the repeat threw each_key_duplicate and the table went blank.
describe("the card search", () => {
  it("lists a card once even if the server sends it twice", async () => {
    vi.useFakeTimers();
    try {
      const row = {
        id: "c1",
        name: "Aang, Swift Savior // Aang and La, Ocean's Fury",
        type_line: "Legendary Creature",
        mana_cost: "{1}{W}{U}",
        set: "tla",
      };
      searchDevCardsMock.mockResolvedValueOnce([
        row,
        row,
        { ...row, id: "c2", name: "Aang's Journey" },
      ]);
      const { container } = mount(false);
      const input = container.querySelector<HTMLInputElement>('input[type="search"]')!;
      input.value = "aang";
      input.dispatchEvent(new Event("input", { bubbles: true }));
      flushSync();
      await vi.advanceTimersByTimeAsync(250);
      flushSync();
      const names = [...container.querySelectorAll(".result .name")].map((n) => n.textContent);
      expect(names).toEqual([row.name, "Aang's Journey"]);
    } finally {
      vi.useRealTimers();
    }
  });
});
