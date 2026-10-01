// readyThemes.test.ts — ADR 0105 §7, sub-PR 6 (#1789): the ready ring
// under the colour-blind palette and the high-contrast theme.
//
// What reaches the DOM is rootSettings.ts (the attributes App.svelte
// writes on :root); what they do is CSS in app.css and Card.svelte. The
// client has no layout engine to evaluate the CSS in, so this reads the
// rules out of those files, the way boardArtPip.test.ts does. If a
// regex here stops matching, the CSS moved: update the pattern, and
// keep the assertions.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, it, expect } from "vitest";

import { applyRootSettings, type RootTarget } from "./rootSettings";
import { defaultSettings } from "./settings";

const source = (rel: string) => readFileSync(fileURLToPath(new URL(rel, import.meta.url)), "utf8");
const appCss = source("../app.css");
const cardSvelte = source("./components/board/Card.svelte");

// block returns the declarations of the rule whose selector is exactly
// `selector`.
function block(css: string, selector: string): string {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const m = new RegExp(`(?:^|\\n)\\s*${escaped} \\{([^}]*)\\}`).exec(css);
  expect(m, `no rule for ${selector}`).not.toBeNull();
  return m![1];
}

function token(decls: string, name: string): string | undefined {
  return new RegExp(`${name}:\\s*([^;]+);`).exec(decls)?.[1].trim();
}

function fakeRoot(): RootTarget & { props: Record<string, string> } {
  const props: Record<string, string> = {};
  return {
    props,
    style: { setProperty: (n, v) => void (props[n] = v) },
    dataset: {},
  };
}

describe("rootSettings: what App.svelte writes on :root", () => {
  it("writes data-colorblind from accessibility.colorblindPalette", () => {
    const s = defaultSettings();
    const root = fakeRoot();
    applyRootSettings(root, s);
    expect(root.dataset.colorblind).toBe("0");
    applyRootSettings(root, {
      ...s,
      accessibility: { ...s.accessibility, colorblindPalette: true },
    });
    expect(root.dataset.colorblind).toBe("1");
  });

  it("still writes everything it wrote before", () => {
    const s = defaultSettings();
    const root = fakeRoot();
    applyRootSettings(root, s);
    expect(root.props["--font-scale"]).toBe(String(s.accessibility.textScale));
    expect(root.dataset.cardSize).toBe(s.display.cardSize);
    expect(root.dataset.tableLayout).toBe(s.display.tableLayout);
    expect(root.dataset.reduceMotion).toBe(s.accessibility.reduceMotion ? "1" : "0");
    expect(root.dataset.alwaysShowFocus).toBe(s.accessibility.alwaysShowFocus ? "1" : "0");
  });

  it("does not apply the theme, which is still inert (App.svelte)", () => {
    const root = fakeRoot();
    applyRootSettings(root, defaultSettings());
    expect(root.dataset.theme).toBeUndefined();
  });
});

describe("the colour-blind ready colour", () => {
  const base = block(appCss, ":root");
  const cb = block(appCss, ':root[data-colorblind="1"]');

  it("swaps --ready and its companions for the alternate set", () => {
    for (const t of ["--ready", "--ready-glow", "--ready-soft", "--ready-ink"]) {
      expect(token(cb, t), t).toBeDefined();
    }
    expect(token(cb, "--ready")).not.toBe(token(base, "--ready"));
  });

  it("is declared after the theme blocks, so it wins under each", () => {
    const at = (sel: string) => appCss.indexOf(`${sel} {`);
    expect(at(':root[data-colorblind="1"]')).toBeGreaterThan(at(':root[data-theme="light"]'));
    expect(at(':root[data-colorblind="1"]')).toBeGreaterThan(
      at(':root[data-theme="high-contrast"]'),
    );
  });

  it("keeps no halo under high contrast", () => {
    const hc = block(appCss, ':root[data-colorblind="1"][data-theme="high-contrast"]');
    expect(token(hc, "--ready-glow")).toBe("transparent");
  });
});

describe("the high-contrast ready ring", () => {
  it("is a solid 3px outline", () => {
    const decls = block(
      cardSvelte,
      ':global(:root[data-theme="high-contrast"]) .card.ready,\n  :global(:root[data-theme="high-contrast"]) .card.ready.combat-target:is(.attacking, .blocking)',
    );
    expect(token(decls, "outline-width")).toBe("3px");
  });

  it("draws no glow", () => {
    const decls = block(
      cardSvelte,
      ':global(:root[data-theme="high-contrast"]) .card.ready::after',
    );
    expect(token(decls, "display")).toBe("none");
    expect(token(block(appCss, ':root[data-theme="high-contrast"]'), "--ready-glow")).toBe(
      "transparent",
    );
  });
});

describe("motion and size (§5, §7)", () => {
  it("the ring does not animate: no transition or animation on .card.ready", () => {
    const ready = block(cardSvelte, ".card.ready");
    const glow = block(cardSvelte, ".card.ready::after");
    for (const d of [ready, glow]) {
      expect(d).not.toMatch(/\btransition\b|\banimation\b/);
    }
  });

  it("pips never draw under 16px", () => {
    expect(token(block(cardSvelte, ".ready-pips"), "--pip")).toMatch(/^max\(16px,/);
  });

  it("at the small card size the counts go and the pip shows alone", () => {
    expect(
      token(block(cardSvelte, ':global(:root[data-card-size="small"]) .pip-count'), "display"),
    ).toBe("none");
  });
});
