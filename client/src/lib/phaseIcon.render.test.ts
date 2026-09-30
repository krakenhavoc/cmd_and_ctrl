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
