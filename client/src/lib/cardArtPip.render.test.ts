// @vitest-environment jsdom
//
// The second #689 proof test, and the other render test the issue named:
// the card-art pip retries on click WITHOUT firing the tile's own click.
//
// `cardArtRetry.ts` owns the state machine and is unit-tested there;
// `cardArt.ts` is "DOM wiring only", and until now nothing tested the
// wiring. The part that can go wrong is exactly the part a pure test
// cannot see: the pip is created inside a tile that plays, taps or
// selects the card on click, so its own click must be stopped before it
// reaches the tile. A regression there does not throw — it plays a land
// every time the player asks for their art back.
//
// This also pins the a11y shape the action chose for that position
// (cardArtRetry.ts's "described" marker): inside a `role="button"`
// tile, whose children assistive tech flattens away, the pip is
// aria-hidden and the tile itself carries the description.

import { describe, it, expect, afterEach, vi } from "vitest";

import Card from "./components/board/Card.svelte";
import { ART_FAILED_TITLE } from "./cardArtRetry";
import type { CardView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

const llanowar = (): CardView =>
  ({
    instance_id: "c1",
    name: "Llanowar Elves",
    owner: "me",
    controller: "me",
    scryfall_id: "aaaa-bbbb",
    type_line: "Creature — Elf Druid",
  }) as unknown as CardView;

interface Tile {
  container: HTMLElement;
  img: HTMLImageElement;
  clicks: CardView[];
  /** The tile's own clickable root. */
  root: HTMLElement;
  pip: () => HTMLElement | null;
}

function mountTile(): Tile {
  // Only setTimeout is faked. Svelte flushes its update batches on a
  // microtask, and faking `queueMicrotask` would leave every render
  // pending behind a timer tick nobody advances.
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });

  const clicks: CardView[] = [];
  const view = render(
    Card as never,
    {
      card: llanowar(),
      onClick: (card: CardView) => clicks.push(card),
    } as never,
  );

  const root = view.container.querySelector<HTMLElement>(".card")!;
  const img = view.container.querySelector<HTMLImageElement>("img")!;
  return {
    container: view.container,
    img,
    clicks,
    root,
    pip: () => view.container.querySelector<HTMLElement>(".card-art-error"),
  };
}

// failTwice walks the art through its whole automatic cycle: the first
// failure, the 2 s wait, the automatic retry, and that retry failing
// too. Only then is the pip supposed to appear.
function failTwice(tile: Tile): void {
  tile.img.dispatchEvent(new Event("error"));
  flushSync();
  expect(tile.pip(), "no pip during the automatic retry window").toBeNull();
  vi.advanceTimersByTime(2000);
  flushSync();
  tile.img.dispatchEvent(new Event("error"));
  flushSync();
}

describe("card-art pip on a clickable board tile", () => {
  it("renders the tile as a control with the art img inside it", () => {
    const tile = mountTile();

    expect(tile.root.getAttribute("role")).toBe("button");
    expect(tile.img.getAttribute("src")).toBe("/cards/aaaa-bbbb/image?size=small");
    expect(tile.pip()).toBeNull();
  });

  it("shows the pip only after the automatic retry has also failed", () => {
    const tile = mountTile();
    failTwice(tile);

    const pip = tile.pip();
    expect(pip, "pip should be showing after the second failure").toBeTruthy();
    expect(pip!.title).toBe(ART_FAILED_TITLE);
    // It sits immediately after the img, which is what the absolute
    // positioning in app.css is anchored on.
    expect(tile.img.nextElementSibling).toBe(pip);
  });

  it("retries the art on click without firing the tile's click", () => {
    const tile = mountTile();
    failTwice(tile);

    // Sanity: the tile really does react to a click of its own, so the
    // assertion below is about propagation and not about a dead tile.
    click(tile.root);
    expect(tile.clicks).toHaveLength(1);
    tile.clicks.length = 0;

    click(tile.pip()!);

    // The retry happened...
    expect(tile.pip()?.classList.contains("busy")).toBe(true);
    // ...and the card was NOT played / tapped / selected by it.
    expect(tile.clicks).toEqual([]);
  });

  it("hides the pip once a retry finally loads", () => {
    const tile = mountTile();
    failTwice(tile);
    expect(tile.pip()).toBeTruthy();

    click(tile.pip()!);
    tile.img.dispatchEvent(new Event("load"));
    flushSync();

    expect(tile.pip()).toBeNull();
    expect(tile.root.hasAttribute("aria-describedby")).toBe(false);
  });

  it("describes the failure on the tile rather than nesting a nameless button", () => {
    const tile = mountTile();
    failTwice(tile);

    const pip = tile.pip()!;
    // Inside a role="button" ancestor a nested button would be a tab
    // stop with no role and no name, so the pip is pointer-only...
    expect(pip.getAttribute("aria-hidden")).toBe("true");
    expect(pip.getAttribute("role")).toBeNull();
    expect(pip.tabIndex).toBe(-1);

    // ...and the tile itself is what announces the failure.
    const descID = tile.root.getAttribute("aria-describedby");
    expect(descID, "the tile should be described by the pip's text").toBeTruthy();
    expect(tile.container.querySelector(`#${descID}`)?.textContent).toBe(ART_FAILED_TITLE);
  });
});
