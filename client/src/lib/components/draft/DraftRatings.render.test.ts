// @vitest-environment jsdom
//
// #5703 (ADR 0222 §7): the rating badge, the credit and the switch, and
// the colour-pair list, rendered.

import { afterEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import type { DraftCard } from "../../draft";
import { clearArchetypeCache } from "../../draftArchetypes";
import { RATINGS_PREF_KEY, type DraftRating } from "../../draftRatings";
import { ratingsPref } from "../../draftRatings.svelte";
import { L } from "../../labels";
import { cleanup, flushSync, render } from "../../test/render.svelte";
import DraftArchetypes from "./DraftArchetypes.svelte";
import DraftArchetypesPanel from "./DraftArchetypesPanel.svelte";
import DraftRatingBadge from "./DraftRatingBadge.svelte";
import DraftRatingsBar from "./DraftRatingsBar.svelte";

const realFetch = globalThis.fetch;

const lands: DraftRating = {
  source: "17lands",
  value: 58.3,
  tier: "B+",
  games: 900,
  set: "trk",
  format: "PremierDraft",
  fetched: "2026-11-11",
};
const house: DraftRating = { source: "c2", value: 3.4 };

const card = (id: string, rating?: DraftRating): DraftCard => ({ id, name: `Card ${id}`, rating });

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

afterEach(() => {
  cleanup();
  ratingsPref.set(true);
  localStorage.clear();
  globalThis.fetch = realFetch;
  clearArchetypeCache();
  vi.restoreAllMocks();
});

const text = (el: Element | null): string => (el?.textContent ?? "").replace(/\s+/g, " ").trim();

describe("DraftRatingBadge", () => {
  it("shows the grade and win rate for 17Lands, with the explanation as its title", () => {
    const { container } = render(
      DraftRatingBadge as never,
      { rating: lands, cardSet: "trk" } as never,
    );
    const b = container.querySelector('[data-testid="rating-badge"]')!;
    expect(b.getAttribute("data-source")).toBe("17lands");
    expect(text(b)).toBe("B+ 58.3%");
    expect(b.getAttribute("title")).toContain("58.3% win rate");
    expect(b.getAttribute("title")).toContain("17Lands");
  });

  it("marks the in-house rating C2, so it is not read as a win rate", () => {
    const { container } = render(DraftRatingBadge as never, { rating: house } as never);
    const b = container.querySelector('[data-testid="rating-badge"]')!;
    expect(b.getAttribute("data-source")).toBe("c2");
    expect(text(b)).toBe("3.4 C2");
    expect(b.getAttribute("title")).toContain("own score");
  });

  it("draws nothing for a card with no rating, or when the viewer hid ratings", () => {
    const none = render(DraftRatingBadge as never, { rating: null } as never);
    expect(none.container.querySelector('[data-testid="rating-badge"]')).toBeNull();
    none.destroy();
    ratingsPref.set(false);
    const hidden = render(DraftRatingBadge as never, { rating: lands } as never);
    expect(hidden.container.querySelector('[data-testid="rating-badge"]')).toBeNull();
  });
});

describe("DraftRatingsBar", () => {
  it("credits 17Lands with a link wherever its numbers show", () => {
    const { container } = render(
      DraftRatingsBar as never,
      { cards: [card("a", lands), card("b", house)] } as never,
    );
    const link = container.querySelector<HTMLAnchorElement>("a")!;
    expect(text(link)).toBe(L.draftRatingsCredit);
    expect(text(link)).toBe("Card ratings from 17Lands");
    expect(link.getAttribute("href")).toBe("https://www.17lands.com/");
    expect(link.getAttribute("rel")).toContain("noopener");
    expect(text(container)).toContain("C2");
  });

  it("owes no credit when only in-house ratings show, and nothing when no card has one", () => {
    const own = render(DraftRatingsBar as never, { cards: [card("b", house)] } as never);
    expect(own.container.querySelector("a")).toBeNull();
    expect(text(own.container)).toContain(L.draftRatingsShow);
    own.destroy();
    const none = render(DraftRatingsBar as never, { cards: [card("c")] } as never);
    expect(none.container.querySelector('[data-testid="ratings-bar"]')).toBeNull();
  });

  it("hides and shows the ratings, and remembers the choice in this browser only", () => {
    const bar = render(DraftRatingsBar as never, { cards: [card("a", lands)] } as never);
    const badge = render(DraftRatingBadge as never, { rating: lands } as never);
    const box = bar.container.querySelector<HTMLInputElement>('input[type="checkbox"]')!;
    expect(box.checked).toBe(true);
    expect(badge.container.querySelector('[data-testid="rating-badge"]')).not.toBeNull();

    box.checked = false;
    box.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    expect(localStorage.getItem(RATINGS_PREF_KEY)).toBe("hide");
    expect(badge.container.querySelector('[data-testid="rating-badge"]')).toBeNull();
    // With the numbers hidden there is nothing to credit, and the switch stays.
    expect(bar.container.querySelector("a")).toBeNull();
    expect(bar.container.querySelector('input[type="checkbox"]')).not.toBeNull();

    box.checked = true;
    box.dispatchEvent(new Event("change", { bubbles: true }));
    flushSync();
    expect(localStorage.getItem(RATINGS_PREF_KEY)).toBeNull();
    expect(badge.container.querySelector('[data-testid="rating-badge"]')).not.toBeNull();
  });
});

describe("DraftArchetypes", () => {
  const body = {
    set: "blb",
    name: "Bloomburrow",
    pairs: [
      {
        colors: "WU",
        label: "Birds and flyers",
        signposts: [
          {
            id: "11111111-1111-1111-1111-111111111111",
            name: "Plumecreed Mentor",
            mana_cost: "{1}{W}{U}",
          },
        ],
      },
      { colors: "UB", label: "Surveil and threshold", signposts: [] },
      { colors: "BR", signposts: [] },
    ],
  };

  it("lists the pairs with their labels and signpost cards, and says where there is none", async () => {
    globalThis.fetch = vi.fn(
      async () =>
        new Response(JSON.stringify(body), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    ) as never;
    const { container } = render(DraftArchetypes as never, { set: "blb" } as never);
    await settle();
    const rows = [...container.querySelectorAll("[data-pair]")];
    expect(rows.map((r) => r.getAttribute("data-pair"))).toEqual(["WU", "UB", "BR"]);
    expect(text(rows[0])).toContain("White-blue");
    expect(text(rows[0])).toContain("Birds and flyers");
    expect(text(rows[0])).toContain("Plumecreed Mentor");
    expect(text(rows[1])).toContain("Surveil and threshold");
    expect(text(rows[2])).toContain("No gold uncommon in this set");
    // A pair with a signpost opens; the card art is asked for only then.
    const details = rows[0].querySelector<HTMLDetailsElement>("details")!;
    expect(details.querySelector("img")).toBeNull();
    details.open = true;
    details.dispatchEvent(new Event("toggle"));
    flushSync();
    expect(details.querySelector("img")?.getAttribute("src")).toBe(
      "/draft/image/11111111-1111-1111-1111-111111111111?size=normal",
    );
    // A pair with no signpost has nothing to open.
    expect(rows[1].querySelector("details")).toBeNull();
  });

  it("says so when the pairs cannot be loaded", async () => {
    globalThis.fetch = vi.fn(
      async () =>
        new Response(JSON.stringify({ error: "nope" }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        }),
    ) as never;
    const { container } = render(DraftArchetypes as never, { set: "blb" } as never);
    await settle();
    expect(text(container)).toContain("Couldn’t load this set’s colour pairs.");
  });
});

describe("DraftArchetypesPanel", () => {
  const body = {
    set: "blb",
    name: "Bloomburrow",
    pairs: [{ colors: "WU", label: "Birds and flyers", signposts: [] }],
  };
  const serve = () => {
    const spy = vi.fn(
      async () =>
        new Response(JSON.stringify(body), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    );
    globalThis.fetch = spy as never;
    return spy;
  };

  it("is a collapsible panel headed by the label, open with the pairs by default", async () => {
    const spy = serve();
    const { container } = render(
      DraftArchetypesPanel as never,
      { set: "blb", setName: "Bloomburrow" } as never,
    );
    await settle();
    const d = container.querySelector<HTMLDetailsElement>('[data-testid="archetypes-panel"]')!;
    expect(d.open).toBe(true);
    expect(text(d.querySelector("summary h2"))).toBe(L.draftArchetypesTitle);
    expect(text(d)).toContain("Bloomburrow");
    expect(text(d)).toContain("Birds and flyers");
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it("asks for nothing while it is closed", async () => {
    const spy = serve();
    const { container } = render(
      DraftArchetypesPanel as never,
      { set: "blb", open: false } as never,
    );
    await settle();
    expect(container.querySelector<HTMLDetailsElement>("details")!.open).toBe(false);
    expect(spy).not.toHaveBeenCalled();
    expect(container.querySelector('[data-testid="archetypes"]')).toBeNull();
  });
});
