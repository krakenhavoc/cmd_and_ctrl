// draftArchetypes.ts — a set's ten colour pairs: types, fetch and the
// small pure helpers the pair list runs on (ADR 0222 §7, amendment
// 2026-11-11, #5703).
//
// Mirrors server/internal/draftpacks/archetypes.go (Archetypes / Pair /
// Signpost). Hand-maintained: update both sides in lockstep.
//
// The route is public, like GET /draft/sets: it is card data.

import type { DraftCard } from "./draft";

/** A gold uncommon of a set: a card that names a colour pair. */
export interface Signpost {
  /** The printing's Scryfall ID: what /draft/image/{id} takes. */
  id: string;
  oracle_id?: string;
  name: string;
  /** Brace notation, "{1}{W}{U}". */
  mana_cost?: string;
}

export interface ArchetypePair {
  /** Two letters in WUBRG order: "WU". */
  colors: string;
  /** A short written label; absent where none is written. */
  label?: string;
  signposts: Signpost[];
}

export interface Archetypes {
  set: string;
  name: string;
  /** The ten pairs, allied first. */
  pairs: ArchetypePair[];
}

const COLOUR_NAMES: Record<string, string> = {
  W: "white",
  U: "blue",
  B: "black",
  R: "red",
  G: "green",
};

/** "WU" → "white-blue". A letter this does not know is left as it is. */
export function pairName(colors: string): string {
  return [...colors].map((c) => COLOUR_NAMES[c] ?? c).join("-");
}

/** "WU" → "White-blue", for a heading. */
export function pairTitle(colors: string): string {
  const n = pairName(colors);
  return n.charAt(0).toUpperCase() + n.slice(1);
}

/** The pair's two letters, as the list of letters a pip row draws. */
export function pairLetters(colors: string): string[] {
  return [...colors];
}

/** True for a pair with a label or a signpost: one worth a row of its own. */
export function pairHasContent(p: ArchetypePair): boolean {
  return !!p.label || p.signposts.length > 0;
}

/**
 * A signpost as the draft card the art component draws. The id is the
 * printing's, so two signposts never share a key.
 */
export function signpostCard(s: Signpost): DraftCard {
  return {
    id: s.id,
    scryfall_id: s.id,
    oracle_id: s.oracle_id,
    name: s.name,
    mana_cost: s.mana_cost,
  };
}

/** The Main picks needed in a colour before the collapsed panel shows its pip (#5848 §4). */
export const RAIL_MIN_PICKS = 3;

/** The WUBRG order the pips and the pairs use. */
const WUBRG = ["W", "U", "B", "R", "G"];

/**
 * How many of `cards` (the Main picks) are in each colour: a gold card
 * counts once for each of its colours, a colourless card or a land not at
 * all.
 */
export function colourCounts(cards: readonly DraftCard[]): Record<string, number> {
  const out: Record<string, number> = {};
  for (const c of cards) for (const k of new Set(c.colors ?? [])) out[k] = (out[k] ?? 0) + 1;
  return out;
}

/**
 * The colours to show as pips on the collapsed panel: each with
 * `RAIL_MIN_PICKS` or more Main picks, most picks first (WUBRG on a tie).
 */
export function railColours(cards: readonly DraftCard[], min = RAIL_MIN_PICKS): string[] {
  const n = colourCounts(cards);
  return WUBRG.filter((k) => (n[k] ?? 0) >= min).sort(
    (a, b) => n[b] - n[a] || WUBRG.indexOf(a) - WUBRG.indexOf(b),
  );
}

/**
 * How many of `cards` fit a colour pair: coloured cards whose colours
 * are all in the pair. A colourless card or a land fits every pair, so
 * counting it would say nothing; it is left out.
 */
export function pairFit(colors: string, cards: readonly DraftCard[]): number {
  const pair = new Set(colors);
  let n = 0;
  for (const c of cards) {
    const k = c.colors ?? [];
    if (k.length > 0 && k.every((x) => pair.has(x))) n++;
  }
  return n;
}

/** The picks that make a pair "a fit": it is highlighted from here. */
export const PAIR_FIT_MIN = 3;

/** The set codes a page asks for are used in a path: only lowercase letters and digits. */
const SET_CODE = /^[a-z0-9]{2,8}$/;

const cache = new Map<string, Promise<Archetypes>>();

/**
 * fetchArchetypes loads a set's colour pairs, once per set per page: the
 * answer changes only with a server build. Plain `fetch`, not
 * `authFetch`: the route is public, and authFetch reads any 401 as an
 * expired session and clears it. A failure is not cached. There is no
 * abort signal: the answer is shared by every caller, so one leaving
 * must not cancel it for the rest.
 */
export function fetchArchetypes(set: string): Promise<Archetypes> {
  const code = set.trim().toLowerCase();
  if (!SET_CODE.test(code)) return Promise.reject(new Error("not a set code"));
  const hit = cache.get(code);
  if (hit) return hit;
  const p = (async () => {
    const res = await fetch(`/draft/sets/${code}/archetypes`, {
      headers: { Accept: "application/json" },
      credentials: "same-origin",
    });
    if (!res.ok) {
      let message = `${res.status} ${res.statusText}`;
      try {
        const body = (await res.json()) as { error?: string };
        if (body.error) message = body.error;
      } catch {
        // not JSON: keep the status line
      }
      throw new Error(message);
    }
    const body = (await res.json()) as Partial<Archetypes>;
    const pairs = Array.isArray(body.pairs) ? body.pairs : [];
    return {
      set: body.set ?? code,
      name: body.name ?? code.toUpperCase(),
      pairs: pairs.map((p) => ({ ...p, signposts: Array.isArray(p.signposts) ? p.signposts : [] })),
    };
  })();
  cache.set(code, p);
  p.catch(() => cache.delete(code));
  return p;
}

/** Forgets what was fetched. For tests. */
export function clearArchetypeCache(): void {
  cache.clear();
}

// Open or closed, kept in localStorage: a per-viewer convenience. Every
// read and write is guarded, because the storage can be missing or refuse
// (a private window, blocked site data); without it the panel is open.

export const ARCHETYPES_OPEN_KEY = "c2.draft.archetypes";

/** Whether the viewer wants the colour-pair panel open. Open unless they closed it. */
export function loadArchetypesOpen(): boolean {
  try {
    return localStorage.getItem(ARCHETYPES_OPEN_KEY) !== "closed";
  } catch {
    return true;
  }
}

/** Remembers the choice; a failure is silent, the choice lasts the page. */
export function saveArchetypesOpen(open: boolean): void {
  try {
    if (open) localStorage.removeItem(ARCHETYPES_OPEN_KEY);
    else localStorage.setItem(ARCHETYPES_OPEN_KEY, "closed");
  } catch {
    // storage unavailable: the choice holds until the page is reloaded
  }
}
