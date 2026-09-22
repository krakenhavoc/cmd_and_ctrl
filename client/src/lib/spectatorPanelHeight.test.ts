// #1071 — spectator panels were capped at the opponent card ceiling
// (200px) even on a tall display, because Board mounts every
// spectator seat with isSelf=false and no `flipped`, which matches
// .panel.opponent's 200px --card-h-max rather than .panel's 240px.
// The reason for that lower ceiling — an opponent panel shares the
// screen with the viewer's own and shouldn't rival it — doesn't apply
// to a spectator, who has no panel of their own.
//
// A real browser is the only thing that resolves `cqh` (a container
// query length) and clamp() against actual layout, so — same posture
// as boardArtPip.test.ts — this reads the numbers out of the source
// files and does the cascade and the arithmetic by hand rather than
// mounting a DOM vitest/jsdom cannot lay out.
//
// If a regex here stops matching, the CSS moved: update the pattern,
// and keep the assertions.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const source = (rel: string) => readFileSync(fileURLToPath(new URL(rel, import.meta.url)), "utf8");
const panelSvelte = source("./components/board/PlayerPanel.svelte");
const boardSvelte = source("./components/board/Board.svelte");

// --- read every `.panel...` rule out of PlayerPanel.svelte's <style> ----
//
// Each rule is keyed by its full class list (["panel"], ["panel",
// "opponent"], ["panel", "opponent", "spectator"], …) so a resolver
// below can replay the cascade for a given element's classes instead
// of hardcoding which rule "wins" — the same kind of tie #1071's
// second comment found by hand when a cross-component :global()
// override turned out to tie with .panel.opponent.flipped on
// specificity.
interface Rule {
  classes: string[];
  props: Record<string, string>;
}

function parsePanelRules(css: string): Rule[] {
  const rules: Rule[] = [];
  const re = /(^|\n)\s*\.panel((?:\.[\w-]+)*)\s*\{([^}]*)\}/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(css))) {
    const classes = ["panel", ...(m[2].match(/[\w-]+/g) ?? [])];
    const props: Record<string, string> = {};
    for (const pm of m[3].matchAll(/(--[\w-]+):\s*([^;]+);/g)) {
      props[pm[1]] = pm[2].trim();
    }
    rules.push({ classes, props });
  }
  expect(rules.length, "no .panel rules parsed — did the CSS move?").toBeGreaterThan(3);
  return rules;
}

const rules = parsePanelRules(panelSvelte);

// resolve replays the cascade for an element carrying `target`'s
// classes: among every rule whose own classes are all present on the
// element and which declares `prop`, the one with the most classes
// wins; a tie goes to whichever is LAST in source order, matching a
// browser's tie-break for equal-specificity author rules. This is
// deliberately not "the rule that mentions `spectator`" — it is the
// same algorithm a browser runs, so a selector that LOOKS like it
// should win but doesn't (the bug this issue's second comment found)
// fails here too.
function resolve(target: string[], prop: string): string {
  const targetSet = new Set(target);
  let best: Rule | null = null;
  for (const r of rules) {
    if (!(prop in r.props)) continue;
    if (!r.classes.every((c) => targetSet.has(c))) continue;
    if (!best || r.classes.length >= best.classes.length) best = r;
  }
  expect(best, `no rule sets ${prop} for .${target.join(".")}`).not.toBeNull();
  return best!.props[prop];
}

function px(value: string): number {
  const m = /^(-?[\d.]+)px$/.exec(value);
  expect(m, `expected a px length, got ${value}`).not.toBeNull();
  return Number(m![1]);
}

// clampCardH evaluates .panel*'s `--card-h: clamp(floor, calc((C cqh
// - O) * scale), var(--card-h-max))` for a target class set, a
// container height in px (1cqh = 1% of it) and a scale factor.
function clampCardH(target: string[], containerHeightPx: number, scale = 1): number {
  const formula = resolve(target, "--card-h");
  const m = /^clamp\((-?[\d.]+)px,\s*calc\(\((-?[\d.]+)cqh\s*-\s*(-?[\d.]+)px\)/.exec(formula);
  expect(m, `could not parse --card-h clamp from ${formula}`).not.toBeNull();
  const [, floorStr, coeffStr, offsetStr] = m!;
  const floor = Number(floorStr);
  const coeff = Number(coeffStr);
  const offset = Number(offsetStr);
  const ceiling = px(resolve(target, "--card-h-max"));
  const raw = coeff * (containerHeightPx / 100) * scale - offset * scale;
  return Math.min(ceiling, Math.max(floor, raw));
}

describe("the .panel --card-h-max ceiling", () => {
  it("is 240px for the self panel, unchanged", () => {
    expect(px(resolve(["panel"], "--card-h-max"))).toBe(240);
  });

  it("is still 200px for an ordinary (non-spectator) opponent panel", () => {
    expect(px(resolve(["panel", "opponent"], "--card-h-max"))).toBe(200);
  });

  it("is still 168px for an ordinary (non-spectator) flipped opponent panel", () => {
    expect(px(resolve(["panel", "opponent", "flipped"], "--card-h-max"))).toBe(168);
  });

  it("is raised to 240px — the self ceiling — for a spectator's opponent panel", () => {
    expect(px(resolve(["panel", "opponent", "spectator"], "--card-h-max"))).toBe(240);
  });

  it("stays raised even if a spectator panel is ever also flipped", () => {
    expect(px(resolve(["panel", "opponent", "flipped", "spectator"], "--card-h-max"))).toBe(240);
  });

  it("leaves the opponent formula's floor and slope untouched", () => {
    // 123px floor, "43cqh - 31px" slope — the same numbers #1071's
    // measurement table cites. Only the ceiling should have moved.
    const formula = resolve(["panel", "opponent", "spectator"], "--card-h");
    expect(formula).toContain("clamp(123px,");
    expect(formula).toContain("43cqh - 31px");
  });
});

describe("the spectator card height at representative panel heights", () => {
  // Content-box heights from #1071's own measurement table (the
  // corrected comment's headless-Chromium numbers), so this test
  // pins the same before/after story a human verified against real
  // layout rather than a number invented for the test.
  it("is unaffected below the old 200px ceiling (483px content height)", () => {
    const before = clampCardH(["panel", "opponent"], 483);
    const after = clampCardH(["panel", "opponent", "spectator"], 483);
    expect(before).toBeCloseTo(176.69, 1);
    expect(after).toBeCloseTo(before, 6);
  });

  it("diverges once the raw slope value passes 200px (605px content height)", () => {
    const before = clampCardH(["panel", "opponent"], 605);
    const after = clampCardH(["panel", "opponent", "spectator"], 605);
    expect(before).toBe(200); // clamped at the old ceiling
    expect(after).toBeCloseTo(229.15, 1); // the raw slope value, under the new ceiling
    expect(after).toBeGreaterThan(before);
  });

  it("reaches the full 240px ceiling on a tall display (663px content height)", () => {
    const before = clampCardH(["panel", "opponent"], 663);
    const after = clampCardH(["panel", "opponent", "spectator"], 663);
    expect(before).toBe(200);
    expect(after).toBe(240);
  });

  it("the two ramps cross at ~537px content height, not before", () => {
    // 43% * 537 - 31 ≈ 199.9px — just under the old ceiling, so
    // before and after still agree one pixel below the crossover.
    const before = clampCardH(["panel", "opponent"], 536);
    const after = clampCardH(["panel", "opponent", "spectator"], 536);
    expect(after).toBeCloseTo(before, 6);
  });
});

describe("the spectator prop reaches PlayerPanel", () => {
  it("PlayerPanel's root element carries the spectator class from the prop", () => {
    expect(panelSvelte).toMatch(/class:spectator\b/);
  });

  it("Board's spectator grid passes spectator={true} to PlayerPanel", () => {
    const spectatorBlock = boardSvelte.split("Spectator path")[1];
    expect(spectatorBlock, "no spectator-path section found in Board.svelte").toBeDefined();
    expect(spectatorBlock).toMatch(/spectator=\{true\}/);
  });

  it("Board's seated (non-spectator) PlayerPanel does not pass spectator={true}", () => {
    const seatedBlock = boardSvelte.split("Spectator path")[0];
    expect(seatedBlock).not.toMatch(/spectator=\{true\}/);
  });
});
