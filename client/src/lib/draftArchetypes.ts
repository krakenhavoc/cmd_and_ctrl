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
