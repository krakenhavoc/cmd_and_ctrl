// place.test.ts — where a hint's card goes (ADR 0125 §3.6, §8): never
// over the anchor, the dock, the stack pile, a dialog or the coach; it
// waits when nothing fits; a phone gets the strip.

import { describe, expect, it } from "vitest";

import { CARD_WIDTH, GAP, overlaps, placeCard, type Placement, type Rect } from "./place";

const VIEW = { width: 1280, height: 800 };
const CARD = { width: CARD_WIDTH, height: 120 };

const rect = (left: number, top: number, width: number, height: number): Rect => ({
  left,
  top,
  width,
  height,
});

function cardRect(p: Placement): Rect {
  if (p.kind !== "card") throw new Error(`expected a card, got ${p.kind}`);
  return rect(p.left, p.top, CARD.width, CARD.height);
}

describe("placeCard", () => {
  it("goes below the anchor when there is room, centred on it", () => {
    const anchor = rect(500, 100, 200, 40);
    const p = placeCard({ anchor, card: CARD, viewport: VIEW });
    expect(p).toEqual({ kind: "card", side: "below", left: 460, top: 140 + GAP });
    expect(overlaps(cardRect(p), anchor)).toBe(false);
  });

  it("goes above an anchor at the bottom of the screen", () => {
    const anchor = rect(500, 700, 200, 60);
    const p = placeCard({ anchor, card: CARD, viewport: VIEW });
    expect(p.kind === "card" && p.side).toBe("above");
    expect(overlaps(cardRect(p), anchor)).toBe(false);
  });

  it("goes beside an anchor that fills the height", () => {
    const anchor = rect(100, 20, 300, 760);
    const p = placeCard({ anchor, card: CARD, viewport: VIEW });
    expect(p.kind === "card" && p.side).toBe("right");
    const left = placeCard({ anchor: rect(880, 20, 300, 760), card: CARD, viewport: VIEW });
    expect(left.kind === "card" && left.side).toBe("left");
  });

  it("never covers the dock, the stack pile, a dialog or the coach", () => {
    const anchor = rect(500, 300, 200, 40);
    // Something to keep clear sits where "below" would go.
    for (const blocker of [
      rect(400, 345, 500, 300), // the dock
      rect(420, 350, 400, 200), // the stack pile
      rect(300, 345, 680, 400), // a dialog
      rect(560, 350, 300, 160), // the coach
    ]) {
      const p = placeCard({ anchor, card: CARD, viewport: VIEW, avoid: [blocker] });
      expect(p.kind === "card" && p.side).toBe("above");
      expect(overlaps(cardRect(p), blocker)).toBe(false);
      expect(overlaps(cardRect(p), anchor)).toBe(false);
    }
  });

  it("keeps clear of every rect it is given at once", () => {
    const anchor = rect(500, 300, 200, 40);
    const avoid = [rect(500, 345, 200, 200), rect(500, 0, 200, 295)];
    const p = placeCard({ anchor, card: CARD, viewport: VIEW, avoid });
    expect(p.kind === "card" && p.side).toBe("right");
    for (const r of [anchor, ...avoid]) expect(overlaps(cardRect(p), r)).toBe(false);
  });

  it("waits when no side fits", () => {
    const anchor = rect(500, 300, 200, 40);
    const p = placeCard({
      anchor,
      card: CARD,
      viewport: VIEW,
      avoid: [
        rect(0, 0, 1280, 295),
        rect(0, 345, 1280, 455),
        rect(0, 300, 495, 40),
        rect(705, 300, 575, 40),
      ],
    });
    expect(p).toEqual({ kind: "wait" });
  });

  it("waits when the card cannot fit on screen at all", () => {
    const p = placeCard({
      anchor: rect(10, 10, 20, 20),
      card: { width: 280, height: 900 },
      viewport: VIEW,
    });
    expect(p).toEqual({ kind: "wait" });
  });

  it("is the one-line strip on a phone, whatever the anchor", () => {
    for (const width of [320, 599]) {
      const p = placeCard({
        anchor: rect(10, 10, 20, 20),
        card: CARD,
        viewport: { width, height: 700 },
      });
      expect(p).toEqual({ kind: "strip" });
    }
    const wide = placeCard({
      anchor: rect(10, 10, 20, 20),
      card: CARD,
      viewport: { width: 600, height: 700 },
    });
    expect(wide.kind).toBe("card");
  });
});
