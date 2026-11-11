// @vitest-environment jsdom
//
// #5437 (ADR 0222 §4): the page lists the sets that can be drafted, names
// the cards the others wait on, and says packs are approximations.

import { afterEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import DraftSets from "./DraftSets.svelte";
import { DRAFT_SETS_TITLE, PACKS_ARE_APPROXIMATIONS } from "../lib/draftSets";
import { L } from "../lib/labels";
import { cleanup, flushSync, render } from "../lib/test/render.svelte";

const realFetch = globalThis.fetch;

function serve(body: unknown, status = 200): ReturnType<typeof vi.fn> {
  const spy = vi.fn(
    async () =>
      new Response(JSON.stringify(body), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
  );
  globalThis.fetch = spy as never;
  return spy;
}

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  vi.restoreAllMocks();
});

const text = (el: Element | null): string => (el?.textContent ?? "").replace(/\s+/g, " ").trim();

const LIST = {
  sets: [
    {
      code: "new",
      name: "Newest Set",
      released: "2026-09-01",
      booster: "play",
      cards: 270,
      draftable: false,
      blockers: ["Card One", "Card Three", "Card Two"],
    },
    {
      code: "rdy",
      name: "Ready Set",
      released: "2026-05-01",
      booster: "play",
      cards: 261,
      draftable: true,
      blockers: [],
    },
    {
      code: "one",
      name: "Almost There",
      released: "2023-02-01",
      booster: "draft",
      cards: 255,
      draftable: false,
      blockers: ["Last Card"],
    },
  ],
  draftable: 1,
};

describe("the draftable sets page", () => {
  it("lists the ready sets, then the others closest first with their blocking cards", async () => {
    const spy = serve(LIST);
    const { container } = render(DraftSets as never, {} as never);
    await settle();

    expect(container.querySelector("h1")?.textContent).toBe(DRAFT_SETS_TITLE);
    // The site header makes calls of its own; the page makes this one.
    expect(spy.mock.calls.map((c) => (c as unknown[])[0])).toContain("/draft/sets");

    const ready = [...container.querySelectorAll('[data-testid="ready-sets"] > li')];
    expect(ready.map((li) => text(li.querySelector(".set-name")))).toEqual(["Ready Set"]);
    expect(text(ready[0])).toContain("rdy");
    expect(text(ready[0])).toContain("2026");
    expect(text(ready[0])).toContain("261 cards");
    // Each set says which booster its packs are modelled on.
    expect(text(ready[0].querySelector(".booster"))).toBe(L.draftPlayBooster);
    expect(text(container.querySelector("#ready-title"))).toBe("Ready to draft 1");

    const blocked = [...container.querySelectorAll('[data-testid="blocked-sets"] > li')];
    expect(blocked.map((li) => text(li.querySelector(".set-name")))).toEqual([
      "Almost There",
      "Newest Set",
    ]);
    expect(text(blocked[0].querySelector(".booster"))).toBe(L.draftDraftBooster);
    expect(text(blocked[1].querySelector(".booster"))).toBe(L.draftPlayBooster);
    expect(text(blocked[0].querySelector(".togo"))).toBe("1 card to go");
    expect(text(blocked[1].querySelector(".togo"))).toBe("3 cards to go");
    expect([...blocked[1].querySelectorAll(".blockers li")].map(text)).toEqual([
      "Card One",
      "Card Three",
      "Card Two",
    ]);
    // The cards sit in a native disclosure, closed until asked for.
    expect(blocked[1].querySelector("details")!.open).toBe(false);
    // A ready set is never shown as waiting on anything.
    expect(ready[0].querySelector("details")).toBeNull();
  });

  it("says packs are approximations of retail packs, before and after the list loads", async () => {
    serve(LIST);
    const { container } = render(DraftSets as never, {} as never);
    const note = () => text(container.querySelector('[data-testid="packs-approximate"]'));
    expect(note()).toBe(PACKS_ARE_APPROXIMATIONS);
    expect(note()).toContain("Packs are approximations of retail packs.");
    // It names both kinds of booster a set can use.
    expect(note()).toContain("play booster");
    expect(note()).toContain("draft booster");
    await settle();
    expect(note()).toBe(PACKS_ARE_APPROXIMATIONS);
  });

  it("offers no draft and no pack to open", async () => {
    serve(LIST);
    const { container } = render(DraftSets as never, {} as never);
    await settle();
    const main = container.querySelector("main")!;
    expect(main.querySelectorAll("button, form, input")).toHaveLength(0);
  });

  it("says so when no set is ready", async () => {
    serve({ sets: [LIST.sets[0]], draftable: 0 });
    const { container } = render(DraftSets as never, {} as never);
    await settle();
    expect(container.querySelector('[data-testid="ready-sets"]')).toBeNull();
    expect(text(container.querySelector("main"))).toContain("No set is ready yet.");
    expect(text(container.querySelector("#ready-title"))).toBe("Ready to draft 0");
  });

  it("shows an empty list from a server with no card data", async () => {
    serve({ sets: [], draftable: 0 });
    const { container } = render(DraftSets as never, {} as never);
    await settle();
    expect(text(container.querySelector("main"))).toContain("No set is ready yet.");
    expect(container.querySelector("#blocked-title")).toBeNull();
    expect(container.querySelector('[role="alert"]')).toBeNull();
  });

  it("reports a failed load", async () => {
    serve({ error: "boom" }, 500);
    const { container } = render(DraftSets as never, {} as never);
    await settle();
    expect(text(container.querySelector('[role="alert"]'))).toContain(
      "Couldn’t load the sets (boom)",
    );
    expect(container.querySelector("#ready-title")).toBeNull();
  });
});
