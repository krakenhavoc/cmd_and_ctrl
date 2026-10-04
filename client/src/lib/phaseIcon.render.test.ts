// @vitest-environment jsdom
//
// #1620: the end step's crescent drew nothing. Its inner arc had a
// radius (4.5) smaller than half the chord it spans (6), which SVG
// silently scales up to exactly half the chord, so the second arc
// retraced the first and the fill had zero area. The pinned dot then
// floated over an empty button. Every step icon has to draw something,
// and no arc may be under-sized for its endpoints.

import { describe, it, expect, afterEach } from "vitest";

import PhaseIcon from "./components/board/PhaseIcon.svelte";
import { STEP_IDS } from "./turn";
import { cleanup, render } from "./test/render.svelte";

afterEach(cleanup);

// Returns the radius shortfall of the worst arc in a path, in SVG
// units (0 when every arc is big enough for its endpoints).
function worstArcShortfall(d: string): number {
  const nums = (s: string) => (s.match(/-?\d*\.?\d+/g) ?? []).map(Number);
  let x = 0;
  let y = 0;
  let worst = 0;
  for (const m of d.matchAll(/([MLHVAZ])([^MLHVAZ]*)/g)) {
    const cmd = m[1];
    const n = nums(m[2]);
    if (cmd === "M" || cmd === "L") [x, y] = [n[0], n[1]];
    else if (cmd === "H") x = n[0];
    else if (cmd === "V") y = n[0];
    else if (cmd === "A") {
      const [rx, ry, , , , ex, ey] = n;
      const half = Math.hypot(ex - x, ey - y) / 2;
      worst = Math.max(worst, half - Math.min(rx, ry));
      [x, y] = [ex, ey];
    }
  }
  return worst;
}

describe("the phase icons", () => {
  for (const step of STEP_IDS) {
    it(`${step} draws a shape`, () => {
      const { container } = render(PhaseIcon as never, { step });
      const shapes = container.querySelectorAll("path, circle, rect, polygon, line");
      expect(shapes.length).toBeGreaterThan(0);
    });

    it(`${step} has no arc too small for its endpoints`, () => {
      const { container } = render(PhaseIcon as never, { step });
      for (const p of container.querySelectorAll("path")) {
        expect(worstArcShortfall(p.getAttribute("d") ?? "")).toBeLessThanOrEqual(1e-9);
      }
    });
  }
});

// #2214: the redrawn set. The two main phases are one glyph, told
// apart only by the numeral on the card; the combat pair are the same
// crossed X, drawn and then sheathed.
describe("the phase icons' pairs (#2214)", () => {
  const svg = (step: string) =>
    render(PhaseIcon as never, { step }).container.querySelector("svg.phase-icon")!;
  const shapesOutside = (el: Element, skip: string) =>
    [...el.querySelectorAll("path, rect, circle, polygon")]
      .filter((n) => !n.closest(skip))
      .map((n) => n.outerHTML);

  it("draws both mains with one glyph, numbered I and II", () => {
    const one = svg("precombat_main");
    const two = svg("postcombat_main");
    expect(shapesOutside(one, ".numeral")).toEqual(shapesOutside(two, ".numeral"));
    expect(one.querySelector(".numeral")?.getAttribute("data-numeral")).toBe("I");
    expect(one.querySelectorAll(".numeral rect")).toHaveLength(1);
    expect(two.querySelector(".numeral")?.getAttribute("data-numeral")).toBe("II");
    expect(two.querySelectorAll(".numeral rect")).toHaveLength(2);
  });

  it("draws begin and end combat as two crossed weapons each", () => {
    for (const step of ["begin_combat", "end_combat"]) {
      expect(svg(step).querySelectorAll(":scope > g")).toHaveLength(2);
    }
  });

  it("gives every step its own drawing", () => {
    const seen = new Map<string, string>();
    for (const step of STEP_IDS) {
      const html = svg(step)
        .innerHTML.replace(/<!--[\s\S]*?-->/g, "")
        .trim();
      const other = seen.get(html);
      expect(other, `${step} draws the same as ${other}`).toBeUndefined();
      seen.set(html, step);
    }
  });

  it("paints in currentColor, so both themes tint it", () => {
    for (const step of STEP_IDS) {
      const el = svg(step);
      expect(el.getAttribute("fill")).toBe("currentColor");
      for (const n of el.querySelectorAll("[fill], [stroke]")) {
        for (const attr of ["fill", "stroke"]) {
          const v = n.getAttribute(attr);
          if (v !== null) expect(["currentColor", "none"]).toContain(v);
        }
      }
    }
  });
});
