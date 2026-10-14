// @vitest-environment jsdom
//
// #3343: a first-use hint's anchor must be ONE element across a page's
// load. The roadmap and catalogue pages render their search box before
// their data lands (so a slow load does not make the hint engine skip the
// page's own tip, #3305), and used to render a second copy once it did:
// the tip's anchor was torn out of the DOM mid-visit, and the hints e2e
// failed with "Element is not attached to the DOM" whenever the fetch was
// slow enough for the tip to attach first. These pin that the box the
// reader (and the tip) holds before the data is the box they hold after.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Catalog from "../routes/Catalog.svelte";
import Roadmap from "../routes/Roadmap.svelte";
import { L } from "./labels";
import { cleanup, flushSync, render } from "./test/render.svelte";

const realFetch = globalThis.fetch;
const realMatchMedia = window.matchMedia;

interface Gate {
  release: () => void;
}

// fetch that answers only when the test lets it, so "before the data" and
// "after the data" are both observable.
function gatedFetch(body: unknown): Gate {
  let release = (): void => {};
  const gate = new Promise<void>((r) => (release = r));
  globalThis.fetch = vi.fn(async () => {
    await gate;
    return new Response(JSON.stringify(body), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  }) as never;
  return { release };
}

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

function matchMedia(narrow: boolean): void {
  window.matchMedia = ((q: string) => ({
    matches: narrow,
    media: q,
    addEventListener: () => {},
    removeEventListener: () => {},
  })) as never;
}

const box = (c: HTMLElement, name: string): HTMLInputElement | null =>
  c.querySelector<HTMLInputElement>(`input[aria-label="${name}"]`);

beforeEach(() => {
  matchMedia(false);
});
afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  window.matchMedia = realMatchMedia;
});

describe("the roadmap's search box", () => {
  const data = {
    counts: {
      implemented: 1,
      partial: 0,
      missing: 0,
      cards: { total: 3, full: 1, caveats: 1, unreviewed: 1 },
    },
    next: [],
    items: [],
  };

  it("is the same element before and after the data lands", async () => {
    const gate = gatedFetch(data);
    const { container } = render(Roadmap as never, {} as never);
    await settle();
    const before = box(container, L.searchRoadmap);
    expect(before, "the box exists while loading").not.toBeNull();
    expect(before!.disabled).toBe(true);
    expect(container.textContent).toContain("Loading the roadmap");

    gate.release();
    await settle();
    const after = box(container, L.searchRoadmap);
    expect(container.textContent).not.toContain("Loading the roadmap");
    expect(after).toBe(before);
    expect(after!.isConnected).toBe(true);
    expect(after!.disabled).toBe(false);
    // And there is only the one.
    expect(container.querySelectorAll(`input[aria-label="${L.searchRoadmap}"]`)).toHaveLength(1);
  });
});

describe("the catalogue's search box", () => {
  const data = {
    counts: { total: 0, full: 0, caveats: 0, unreviewed: 0, missing: 0 },
    cards: [],
  };

  for (const narrow of [false, true]) {
    it(`is the same element before and after the data lands (${narrow ? "narrow" : "wide"})`, async () => {
      matchMedia(narrow);
      const gate = gatedFetch(data);
      const { container } = render(Catalog as never, {} as never);
      await settle();
      const before = box(container, L.searchCatalogue);
      expect(before, "the box exists while loading").not.toBeNull();
      expect(before!.disabled).toBe(true);
      expect(container.textContent).toContain("Loading the catalogue");

      gate.release();
      await settle();
      const after = box(container, L.searchCatalogue);
      expect(container.textContent).not.toContain("Loading the catalogue");
      expect(after).toBe(before);
      expect(after!.isConnected).toBe(true);
      expect(after!.disabled).toBe(false);
      expect(container.querySelectorAll(`input[aria-label="${L.searchCatalogue}"]`)).toHaveLength(
        1,
      );
    });
  }
});
