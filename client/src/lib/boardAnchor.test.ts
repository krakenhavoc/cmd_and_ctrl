// @vitest-environment jsdom
//
// boardAnchor.test.ts — ADR 0120 §3: every lookup of a card tile or a
// seat's avatar by `data-instance-id` / `data-seat-id` goes through one
// helper. With no expanded overlay mounted it is the old first match, so
// nothing on the table moves; with one mounted the overlay's copy wins,
// for the helper and for each reader that goes through it.

import { afterEach, describe, expect, it } from "vitest";
import {
  BOARD_EXPANDED_ATTR,
  cardAnchors,
  cardSelector,
  findAnchor,
  findCardAnchor,
  findSeatAnchor,
  hasSize,
  seatSelector,
} from "./boardAnchor";
import { findTarget } from "./stackArrows";
import { resolveAnchor } from "./tutorialAnchor";

afterEach(() => {
  document.body.innerHTML = "";
});

// jsdom lays nothing out, so every rect is 0×0. A test gives an element
// a size by stubbing its rect.
function sized(el: Element, w = 100, h = 140): void {
  el.getBoundingClientRect = () =>
    ({ left: 0, top: 0, right: w, bottom: h, width: w, height: h, x: 0, y: 0 }) as DOMRect;
}

// A board with two seats, each with an avatar and a card; `expanded`
// adds an overlay after the slots (as ADR 0120 §2 mounts it) holding a
// second copy of bob's board.
function board(expanded: boolean): HTMLElement {
  const el = document.createElement("div");
  el.className = "board";
  el.innerHTML = `
    <div class="slot" data-pos="self">
      <div data-seat-id="me" data-copy="table"></div>
      <div class="card" data-instance-id="mine" data-copy="table"></div>
    </div>
    <div class="slot">
      <div data-seat-id="bob" data-copy="table"></div>
      <div class="card" data-instance-id="bear" data-copy="table"></div>
    </div>
    ${
      expanded
        ? `<section ${BOARD_EXPANDED_ATTR}>
             <div data-seat-id="bob" data-copy="overlay"></div>
             <div class="card" data-instance-id="bear" data-copy="overlay"></div>
           </section>`
        : ""
    }`;
  document.body.appendChild(el);
  for (const n of el.querySelectorAll("[data-copy]")) sized(n);
  return el;
}

const copyOf = (el: Element | null) => (el as HTMLElement | null)?.dataset.copy ?? null;

describe("findAnchor", () => {
  it("is the first match in document order with no overlay mounted", () => {
    const b = board(false);
    // A second table copy later in the DOM loses, as querySelector's did.
    const later = document.createElement("div");
    later.dataset.instanceId = "bear";
    later.dataset.copy = "later";
    b.appendChild(later);
    expect(copyOf(findCardAnchor(b, "bear"))).toBe("table");
    expect(copyOf(findSeatAnchor(b, "bob"))).toBe("table");
    expect(findCardAnchor(b, "bear")).toBe(b.querySelector(cardSelector("bear")));
  });

  it("prefers the overlay's copy while one is mounted, though it comes later", () => {
    const b = board(true);
    expect(copyOf(findCardAnchor(b, "bear"))).toBe("overlay");
    expect(copyOf(findSeatAnchor(b, "bob"))).toBe("overlay");
  });

  it("falls back to the table for anything the overlay does not draw", () => {
    const b = board(true);
    expect(copyOf(findCardAnchor(b, "mine"))).toBe("table");
    expect(copyOf(findSeatAnchor(b, "me"))).toBe("table");
  });

  it("goes back to the table copy when the overlay closes", () => {
    const b = board(true);
    b.querySelector(`[${BOARD_EXPANDED_ATTR}]`)!.remove();
    expect(copyOf(findCardAnchor(b, "bear"))).toBe("table");
  });

  it("finds the overlay from the document root too", () => {
    board(true);
    expect(copyOf(findCardAnchor(document, "bear"))).toBe("overlay");
  });

  it("skips an element `accept` refuses, such as one with no size", () => {
    const b = board(true);
    const overlayCopy = b.querySelector(`[${BOARD_EXPANDED_ATTR}] ${cardSelector("bear")}`)!;
    sized(overlayCopy, 0, 0);
    expect(hasSize(overlayCopy)).toBe(false);
    expect(copyOf(findCardAnchor(b, "bear", { accept: hasSize }))).toBe("table");
    // Without `accept`, a match is a match.
    expect(copyOf(findCardAnchor(b, "bear"))).toBe("overlay");
  });

  it("is null when nothing matches or nothing is accepted", () => {
    const b = board(true);
    expect(findCardAnchor(b, "nobody")).toBeNull();
    expect(findAnchor(b, seatSelector("bob"), { accept: () => false })).toBeNull();
  });

  it("escapes the ID it is given", () => {
    const b = board(false);
    const odd = document.createElement("div");
    odd.dataset.instanceId = 'a"b';
    b.appendChild(odd);
    expect(findCardAnchor(b, 'a"b')).toBe(odd);
  });
});

describe("cardAnchors", () => {
  it("is one element per ID, the one findCardAnchor returns", () => {
    const b = board(true);
    const all = cardAnchors(b);
    expect([...all.keys()].sort()).toEqual(["bear", "mine"]);
    expect(copyOf(all.get("bear")!)).toBe("overlay");
    expect(copyOf(all.get("mine")!)).toBe("table");
    for (const [id, el] of all) expect(el).toBe(findCardAnchor(b, id));
  });

  it("is the first match per ID with no overlay", () => {
    const b = board(false);
    expect(copyOf(cardAnchors(b).get("bear")!)).toBe("table");
  });
});

describe("the readers go through it", () => {
  it("the fan lane's findTarget picks the overlay copy, and still skips the lane", () => {
    const b = board(true);
    const lane = document.createElement("div");
    lane.innerHTML = `<div data-instance-id="bear" data-copy="lane"></div>`;
    sized(lane.firstElementChild!);
    b.prepend(lane);
    expect(copyOf(findTarget(b, lane, "permanent", "bear"))).toBe("overlay");
    expect(copyOf(findTarget(b, lane, "player", "bob"))).toBe("overlay");
    b.querySelector(`[${BOARD_EXPANDED_ATTR}]`)!.remove();
    expect(copyOf(findTarget(b, lane, "permanent", "bear"))).toBe("table");
  });

  it("a tutorial spotlight anchors on the overlay copy", () => {
    board(true);
    expect(copyOf(resolveAnchor({ cardID: "bear" }))).toBe("overlay");
    expect(copyOf(resolveAnchor({ seatID: "bob" }))).toBe("overlay");
    expect(copyOf(resolveAnchor({ seatID: "me" }))).toBe("table");
  });
});
