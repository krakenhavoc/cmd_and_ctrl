// index.ts — every first-use hint, collected (ADR 0125 §3.2).
//
// Each hint is a file `<Feature>.hint.ts` beside the component that owns
// the feature, with a `default` export of one Hint. There is no central
// list: two PRs that add hints never conflict, and the PR that changes a
// component shows its hint right beside it. Vite (and vitest) gather the
// files at build time.
//
// The pattern covers all of client/src, routes included (the ADR's
// `"../**/*.hint.ts"` would only reach lib/). Test fixtures under
// lib/test/ never ship; hints.test.ts collects them on their own.
//
// A removed hint's id goes in retired.ts, never back into use.

import type { Hint } from "./hint";

type HintModule = { default?: unknown };

/** isHint is the shape check a collected module's default export must pass. */
function isHint(v: unknown): v is Hint {
  if (!v || typeof v !== "object") return false;
  const h = v as Partial<Hint>;
  return (
    typeof h.id === "string" &&
    typeof h.version === "number" &&
    typeof h.place === "string" &&
    typeof h.order === "number" &&
    h.anchor !== undefined &&
    h.title !== undefined &&
    h.body !== undefined
  );
}

/**
 * collectHints turns a glob of hint modules into the hints they export.
 * A module whose default export is not a hint is a build-time mistake
 * and throws, naming the file.
 */
export function collectHints(modules: Record<string, HintModule>): Hint[] {
  const out: Hint[] = [];
  for (const [path, mod] of Object.entries(modules).sort(([a], [b]) => (a < b ? -1 : 1))) {
    if (!isHint(mod.default)) {
      throw new Error(`hints: ${path} must have a default export of one Hint`);
    }
    out.push(mod.default);
  }
  return out;
}

/** HINTS is every hint the client ships. */
export const HINTS: readonly Hint[] = collectHints(
  import.meta.glob<HintModule>(["/src/**/*.hint.ts", "!/src/lib/test/**"], { eager: true }),
);
