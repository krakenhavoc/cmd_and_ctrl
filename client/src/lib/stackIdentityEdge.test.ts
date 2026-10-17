// ADR 0119 amendment 2026-10-16 §C — the pile's top item in its
// colours: the WUBRG read off the card view, and the hard-banded edge
// (mono, two colours, five, colourless, and none for a hidden card).

import { describe, expect, it } from "vitest";

import { identityEdgeBackground, identityLetters, identityTokens } from "./stackIdentityEdge";

const CARD = 488 / 680;

/** The stops of a conic edge, as [token, from°, to°]. */
function stops(css: string): [string, number, number][] {
  const body = css.replace(/^conic-gradient\(from [\d.]+deg, /, "").replace(/\)$/, "");
  return body.split(/, (?=var\()/).map((s) => {
    const m = /^(var\(--identity-[a-z]\)) ([\d.]+)deg ([\d.]+)deg$/.exec(s);
    if (!m) throw new Error(`bad stop: ${s}`);
    return [m[1], Number(m[2]), Number(m[3])];
  });
}

/** The bearing (clockwise from up) of a point on a w×1 box's rim. */
function bearing(x: number, y: number): number {
  const a = (Math.atan2(x, -y) * 180) / Math.PI;
  return a < 0 ? a + 360 : a;
}

describe("identityLetters", () => {
  it("reads the card's colours in WUBRG order, whatever order they come in", () => {
    expect(identityLetters({ colors: ["G", "W", "R"] })).toEqual(["W", "R", "G"]);
  });

  it("is [] for a colourless card (colors omitted) and null with no card to read", () => {
    expect(identityLetters({})).toEqual([]);
    expect(identityLetters({ colors: [] })).toEqual([]);
    expect(identityLetters(null)).toBeNull();
    expect(identityLetters(undefined)).toBeNull();
  });

  it("names only the fixed identity tokens", () => {
    expect(identityTokens(["U", "R"])).toEqual(["var(--identity-u)", "var(--identity-r)"]);
    expect(identityTokens([])).toEqual(["var(--identity-c)"]);
  });
});

describe("identityEdgeBackground", () => {
  it("is one colour all round for a mono item", () => {
    expect(identityEdgeBackground({ colors: ["R"] }, CARD)).toBe("var(--identity-r)");
  });

  it("is --identity-c for a colourless item", () => {
    expect(identityEdgeBackground({ colors: [] }, CARD)).toBe("var(--identity-c)");
    expect(identityEdgeBackground({}, CARD)).toBe("var(--identity-c)");
  });

  it("is nothing for a hidden source or face-down spell (no card view)", () => {
    expect(identityEdgeBackground(null, CARD)).toBeNull();
  });

  it("starts at the top-left corner and splits two colours at the opposite corner", () => {
    const css = identityEdgeBackground({ colors: ["R", "U"] }, CARD)!;
    const start = bearing(-CARD / 2, -0.5);
    expect(css.startsWith(`conic-gradient(from ${Math.round(start * 100) / 100}deg, `)).toBe(true);
    const s = stops(css);
    expect(s.map((x) => x[0])).toEqual(["var(--identity-u)", "var(--identity-r)"]);
    // Half the rim clockwise from the top-left corner is the bottom-right
    // corner: blue runs the top and the right side, red the bottom and left.
    const half = bearing(CARD / 2, 0.5) - start + 360;
    expect(s[0][1]).toBe(0);
    expect(s[0][2]).toBeCloseTo(half, 1);
    expect(s[1]).toEqual([s[1][0], s[0][2], 360]);
  });

  it("gives five colours equal runs of the rim, W U B R G clockwise, with hard stops", () => {
    const css = identityEdgeBackground({ colors: ["G", "R", "B", "U", "W"] }, CARD)!;
    const s = stops(css);
    expect(s.map((x) => x[0])).toEqual([
      "var(--identity-w)",
      "var(--identity-u)",
      "var(--identity-b)",
      "var(--identity-r)",
      "var(--identity-g)",
    ]);
    // Hard bands: each starts exactly where the last ends, round to 360.
    for (let i = 1; i < s.length; i++) expect(s[i][1]).toBe(s[i - 1][2]);
    expect(s[0][1]).toBe(0);
    expect(s[4][2]).toBe(360);
    // Each boundary is the bearing of a point k/5 of the way round the
    // rim, so the runs are equal on the card, not equal angles.
    const per = 2 * (CARD + 1);
    const start = bearing(-CARD / 2, -0.5);
    // 1/5 of the rim is short of the top edge's end (the top is the
    // shortest side): still on the top edge.
    const p1 = per / 5;
    expect(p1).toBeLessThan(CARD);
    const b1 = (bearing(-CARD / 2 + p1, -0.5) - start + 360) % 360;
    expect(s[0][2]).toBeCloseTo(b1, 1);
    // Angles differ, runs do not.
    const widths = s.map(([, a, b]) => b - a);
    expect(Math.max(...widths) - Math.min(...widths)).toBeGreaterThan(5);
  });

  it("follows the box it is drawn on: an ability card's 63:88", () => {
    const a = identityEdgeBackground({ colors: ["W", "B"] }, 63 / 88)!;
    const b = identityEdgeBackground({ colors: ["W", "B"] }, CARD)!;
    expect(a).not.toBe(b);
    expect(stops(a).map((x) => x[0])).toEqual(["var(--identity-w)", "var(--identity-b)"]);
  });
});
