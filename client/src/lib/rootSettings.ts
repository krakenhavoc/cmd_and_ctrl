// rootSettings.ts — the settings that hang off :root, as one function.
//
// App.svelte calls it from an $effect whenever the settings store
// changes. It lives here, rather than inline in that effect, so a test
// can hold the attribute names the stylesheet keys on (app.css's
// :root[data-…] rules) without mounting the whole app.

import type { Settings } from "./settings";
import { ACCENT_VARS, accentVars } from "./skins";

/** The part of an element applyRootSettings writes to. */
export interface RootTarget {
  style: { setProperty(name: string, value: string): void; removeProperty(name: string): void };
  dataset: Record<string, string | undefined>;
}

/**
 * applyRootSettings writes the stylesheet-driving settings onto `root`
 * (document.documentElement in the app).
 *
 * `display.theme` is the skin: `data-theme` picks its token block in
 * app.css (lib/skins.ts). `display.accent`, when set, overrides the
 * skin's accent family inline; cleared, the inline values go and the
 * skin's own accent shows again.
 *
 * `data-colorblind` (ADR 0105 §7, sub-PR 6) is the first consumer of
 * `accessibility.colorblindPalette`: app.css swaps the `--ready` tokens
 * for the alternate set under it. The seat palette the setting is
 * named for does not read it yet.
 */
export function applyRootSettings(root: RootTarget, s: Settings): void {
  root.style.setProperty("--font-scale", String(s.accessibility.textScale));
  root.dataset.cardSize = s.display.cardSize;
  root.dataset.tableLayout = s.display.tableLayout;
  root.dataset.reduceMotion = s.accessibility.reduceMotion ? "1" : "0";
  root.dataset.alwaysShowFocus = s.accessibility.alwaysShowFocus ? "1" : "0";
  root.dataset.colorblind = s.accessibility.colorblindPalette ? "1" : "0";
  root.dataset.theme = s.display.theme;
  if (s.display.accent) {
    for (const [name, value] of Object.entries(accentVars(s.display.accent))) {
      root.style.setProperty(name, value);
    }
  } else {
    for (const name of ACCENT_VARS) root.style.removeProperty(name);
  }
}
