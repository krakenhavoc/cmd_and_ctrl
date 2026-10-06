// @vitest-environment jsdom
//
// hoverHit.render.test.ts — #2396. A hovered card must not move the box
// the pointer is on: when it did, a pointer resting near the card's
// bottom edge was left behind as the card rose, the card lost the
// hover, dropped back and rose again, over and over. jsdom lays nothing
// out and has no :hover, so the pointer itself is checked in the browser
// (tests-e2e/tests/hover-hit-2396.spec.ts). This pins the structure that
// makes it hold: what lifts, and what stays put.

import { describe, it, expect, afterEach, vi } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

vi.mock("./sounds", () => ({ play: () => {} }));

import Hand from "./components/board/Hand.svelte";
import type { CardView } from "./protocol";
import { innerLift } from "./dragCast";
import { render, cleanup, flushSync } from "./test/render.svelte";

const src = (p: string) => readFileSync(join(process.cwd(), p), "utf8");
// The declarations of the first rule whose selector line is exactly `sel`.
const rule = (css: string, sel: string): string => {
  const at = css.indexOf(`\n${sel} {`);
  if (at < 0) throw new Error(`no rule ${sel}`);
  return css.slice(at).split("}")[0];
};

afterEach(() => cleanup());

describe("a hovered card's box stays where it rests (#2396)", () => {
  it("Card grows from its bottom edge instead of moving up", () => {
    const css = src("src/lib/components/board/Card.svelte");
    const card = rule(css, "  .card");
    // The scale's origin moved to the bottom centre, around the tap turn.
    expect(card).toContain(
      "transform: rotate(var(--tap-rot, 0deg)) translateY(50%) scale(var(--hover-scale, 1))",
    );
    expect(card).toContain("translateY(-50%)");
    const hover = rule(css, "  .card.clickable:hover");
    expect(hover).toContain("--hover-scale: 1.04");
    expect(css).not.toContain("--hover-lift");
  });

  it("the hand lifts its cards inside slots that stay put", () => {
    const css = src("src/lib/components/board/Hand.svelte");
    expect(rule(css, "  .hand:not(.opponent):hover")).not.toContain("transform");
    expect(rule(css, "  .hand:not(.opponent):not(.scroll):hover > .hand-slot > .rise")).toContain(
      "translateY(var(--hand-rise))",
    );
  });

  it("the castable strip lifts its cards inside slots that stay put", () => {
    const css = src("src/lib/components/board/ExileStrip.svelte");
    const row = rule(css, "  .strip-cards:hover");
    // Neither moved nor grown: the strip aligns the row to its bottom
    // edge, so a row grown to a whole card would carry its slots up.
    expect(row).not.toContain("transform");
    expect(row).not.toContain("max-height");
    expect(rule(css, "  .strip-slot.castable:hover")).not.toContain("transform");
    expect(rule(css, "  .strip-cards:hover .rise")).toContain("translateY(var(--strip-rise))");
  });

  it("an opening-hand card lifts its face and stays in the arc", () => {
    const css = src("src/routes/Game.svelte");
    expect(rule(css, "  .mulligan-card:hover")).not.toContain("transform");
    expect(rule(css, "  .mulligan-card:hover > .mulligan-face")).toContain(
      "translateY(-22px) scale(1.12)",
    );
  });

  it("each hand slot holds a .rise around its dealt card, and publishes its tilt", () => {
    const cards: CardView[] = [0, 1, 2].map((i) => ({
      instance_id: `c${i}`,
      name: `Card ${i}`,
      owner: "me",
      controller: "me",
    }));
    const { container } = render(
      Hand as never,
      {
        hand: { kind: "hand", owner: "me", count: 3, cards },
        isSelf: true,
        viewerID: "me",
      } as never,
    );
    flushSync();
    const slots = [...container.querySelectorAll<HTMLElement>(".hand-slot")];
    expect(slots).toHaveLength(3);
    for (const s of slots) {
      expect(s.querySelector(":scope > .rise > .deal-wrap .card")).not.toBeNull();
    }
    expect(slots[0].style.getPropertyValue("--slot-rot")).toBe("-7deg");
    expect(slots[2].style.getPropertyValue("--slot-rot")).toBe("7deg");
  });
});

describe("innerLift", () => {
  const box = (top: number): DOMRect =>
    ({ top, left: 0, right: 10, bottom: top + 10, width: 10, height: 10 }) as DOMRect;

  it("is how far a slot's .rise sits above the slot", () => {
    const slot = document.createElement("div");
    const rise = document.createElement("div");
    rise.className = "rise";
    slot.appendChild(rise);
    slot.getBoundingClientRect = () => box(500);
    rise.getBoundingClientRect = () => box(360);
    expect(innerLift(slot)).toBe(140);
    rise.getBoundingClientRect = () => box(500);
    expect(innerLift(slot)).toBe(0);
  });

  it("is 0 with no .rise", () => {
    expect(innerLift(document.createElement("div"))).toBe(0);
  });
});
