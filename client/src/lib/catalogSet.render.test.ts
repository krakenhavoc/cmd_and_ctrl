// @vitest-environment jsdom
//
// #4790: the catalogue's set picker narrows the grid to one set and states
// how much of that set the game plays. The filter and the summary line have
// unit tests in catalog.test.ts; this pins that the page wires them: the
// picker is there, picking a set shows only its cards, and the line under
// the bar says the counts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Catalog from "../routes/Catalog.svelte";
import { L } from "./labels";
import { cleanup, flushSync, render } from "./test/render.svelte";

const realFetch = globalThis.fetch;
const realMatchMedia = window.matchMedia;

const data = {
  counts: { total: 2, full: 2, caveats: 0, unreviewed: 0, missing: 0 },
  cards: [
    { oracle_id: "a", name: "Web Slinger", completeness: "full", sets: ["spm"] },
    { oracle_id: "b", name: "Old Bear", completeness: "full", sets: ["old"] },
  ],
  sets: [
    {
      code: "spm",
      name: "Marvel's Spider-Man",
      released: "2026-09-26",
      total: 3,
      full: 1,
      caveats: 0,
      unreviewed: 0,
      unsupported: 2,
    },
    {
      code: "old",
      name: "Old Set",
      released: "2001-01-01",
      total: 1,
      full: 1,
      caveats: 0,
      unreviewed: 0,
      unsupported: 0,
    },
  ],
};

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => {
  window.matchMedia = ((q: string) => ({
    matches: false,
    media: q,
    addEventListener: () => {},
    removeEventListener: () => {},
  })) as never;
  globalThis.fetch = vi.fn(
    async () =>
      new Response(JSON.stringify(data), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
  ) as never;
});
afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  window.matchMedia = realMatchMedia;
});

describe("the catalogue's set picker", () => {
  it("lists the sets, filters the grid to the picked one and states its counts", async () => {
    const { container } = render(Catalog as never, {} as never);
    await settle();

    const select = container.querySelector<HTMLSelectElement>(
      `select[aria-label="${L.filterCatalogueBySet}"]`,
    );
    expect(select, "the picker is drawn").not.toBeNull();
    expect([...select!.options].map((o) => o.textContent)).toEqual([
      "All sets",
      "Marvel's Spider-Man (1/3)",
      "Old Set (1/1)",
    ]);
    expect(container.textContent).toContain("Web Slinger");
    expect(container.textContent).toContain("Old Bear");
    expect(container.textContent).not.toContain("not yet supported");

    select!.value = "spm";
    select!.dispatchEvent(new Event("change", { bubbles: true }));
    await settle();

    expect(container.textContent).toContain("Web Slinger");
    expect(container.textContent).not.toContain("Old Bear");
    expect(container.textContent).toContain(
      "Marvel's Spider-Man: 1 of 3 supported — 1 complete; 2 not yet supported",
    );
  });

  it("is not drawn when the server sent no sets", async () => {
    globalThis.fetch = vi.fn(
      async () =>
        new Response(JSON.stringify({ counts: data.counts, cards: data.cards }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    ) as never;
    const { container } = render(Catalog as never, {} as never);
    await settle();
    expect(container.querySelector(`select[aria-label="${L.filterCatalogueBySet}"]`)).toBeNull();
    expect(container.textContent).toContain("Web Slinger");
  });
});
