// catalog.ts — the public card catalogue: types, fetch, and the
// filtering the page runs over a few hundred cards.
//
// Mirrors server/internal/catalog/catalog.go (Entry / Face / Counts /
// Response). Hand-maintained, like protocol.ts — update both sides in
// lockstep.
//
// The filter lives here rather than in Catalog.svelte for the reason
// contextMenu.logic.ts and zoneBrowser.logic.ts do: decisions a user
// can get wrong deserve unit tests, and a `.svelte` file cannot have
// them (there is no jsdom in this project).

/** How faithfully the engine implements a card's printed text. */
export type Completeness = "full" | "caveats" | "unreviewed";

export interface CatalogFace {
  index: number;
  name: string;
  type_line?: string;
  oracle_text?: string;
  mana_cost?: string;
  /** Whether the registry has a spec for this specific face. */
  automated: boolean;
}

export interface CatalogEntry {
  oracle_id: string;
  /** Absent when the loaded Scryfall dump has no printing (see `missing`). */
  scryfall_id?: string;
  name: string;
  mana_cost?: string;
  type_line?: string;
  oracle_text?: string;
  /** Lowercased card types, parsed server-side. */
  types?: string[];
  /** WUBRG letters. Empty array means colourless. */
  color_identity?: string[];
  colors?: string[];
  completeness: Completeness;
  /** Present exactly when completeness is "caveats". */
  caveats?: string[];
  faces?: CatalogFace[];
  /** The catalog registers this card but the dump has no printing for it. */
  missing?: boolean;
}

export interface CatalogCounts {
  total: number;
  full: number;
  caveats: number;
  unreviewed: number;
  missing: number;
}

export interface CatalogResponse {
  counts: CatalogCounts;
  cards: CatalogEntry[];
}

/**
 * fetchCatalog loads the published catalogue.
 *
 * Plain `fetch`, not `authFetch`: the route is public, and authFetch
 * reads any 401 as an expired session and clears it — the same reason
 * previewGame in api.ts opts out. Nobody should be logged out by
 * visiting a showcase page.
 */
export async function fetchCatalog(signal?: AbortSignal): Promise<CatalogResponse> {
  const res = await fetch("/catalog", {
    headers: { Accept: "application/json" },
    credentials: "same-origin",
    signal,
  });
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // not JSON — keep the status line
    }
    throw new Error(message);
  }
  return (await res.json()) as CatalogResponse;
}

/** The five colours plus the colourless bucket, in WUBRG order. */
export const COLOR_FILTERS = ["W", "U", "B", "R", "G", "C"] as const;
export type ColorFilter = (typeof COLOR_FILTERS)[number];

export const COLOR_LABELS: Record<ColorFilter, string> = {
  W: "White",
  U: "Blue",
  B: "Black",
  R: "Red",
  G: "Green",
  C: "Colourless",
};

/** Card types offered by the type filter, in the order they appear. */
export const TYPE_FILTERS = [
  "creature",
  "instant",
  "sorcery",
  "artifact",
  "enchantment",
  "land",
  "planeswalker",
  "battle",
] as const;
export type TypeFilter = (typeof TYPE_FILTERS)[number];

export interface CatalogFilter {
  /** Free text matched against name, type line and oracle text. */
  query: string;
  /** Empty means "any colour". Multiple selections are OR'd. */
  colors: ColorFilter[];
  /** Empty means "any type". Multiple selections are OR'd. */
  types: TypeFilter[];
  /** Empty means "any". Multiple selections are OR'd. */
  completeness: Completeness[];
}

export const EMPTY_FILTER: CatalogFilter = {
  query: "",
  colors: [],
  types: [],
  completeness: [],
};

/**
 * matchesColor tests one card against one colour selection.
 *
 * Filtering is on COLOR IDENTITY, not the card's colours: this is a
 * Commander engine, identity is the question players actually ask,
 * and it gives lands and mana rocks a sensible answer where `colors`
 * is empty. "C" therefore means "no colour identity at all" — Sol
 * Ring qualifies, a Signet does not.
 */
export function matchesColor(entry: CatalogEntry, color: ColorFilter): boolean {
  const identity = entry.color_identity ?? [];
  if (color === "C") return identity.length === 0;
  return identity.includes(color);
}

/**
 * matchesQuery tests a card against free text.
 *
 * Matches name, type line and oracle text, plus every printed face's
 * name and text so searching "Akoum Teeth" finds the card whose front
 * is called something else. Case-insensitive substring; no ranking,
 * because the result set is already sorted by name and a few hundred
 * cards do not need relevance ordering.
 */
export function matchesQuery(entry: CatalogEntry, query: string): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  const haystack = [
    entry.name,
    entry.type_line ?? "",
    entry.oracle_text ?? "",
    ...(entry.faces ?? []).flatMap((f) => [f.name, f.type_line ?? "", f.oracle_text ?? ""]),
  ]
    .join("\n")
    .toLowerCase();
  return haystack.includes(q);
}

/**
 * filterCatalog applies the whole filter. Within a facet the
 * selections are OR'd; across facets they are AND'd, which is what a
 * player means by "red creatures that fully work".
 */
export function filterCatalog(entries: CatalogEntry[], filter: CatalogFilter): CatalogEntry[] {
  return entries.filter((e) => {
    if (!matchesQuery(e, filter.query)) return false;
    if (filter.colors.length > 0 && !filter.colors.some((c) => matchesColor(e, c))) {
      return false;
    }
    if (filter.types.length > 0) {
      const types = e.types ?? [];
      if (!filter.types.some((t) => types.includes(t))) return false;
    }
    if (filter.completeness.length > 0 && !filter.completeness.includes(e.completeness)) {
      return false;
    }
    return true;
  });
}

/** toggle adds or removes one value from a filter facet. */
export function toggle<T>(list: T[], value: T): T[] {
  return list.includes(value) ? list.filter((v) => v !== value) : [...list, value];
}

/**
 * catalogImageURL builds the URL for a card's art.
 *
 * NOT scryfallImageURL from cardImage.ts: that points at
 * /cards/<id>/image, which requires a session, and this page is
 * public. The catalogue has its own route scoped to catalogue cards —
 * see server/internal/catalog/http.go.
 *
 * Face 0 omits the `face` parameter, exactly as cardImage.ts does:
 * emitting `face=0` would miss every existing cache entry (service
 * worker, browser and server disk all key on the URL).
 */
export function catalogImageURL(
  entry: CatalogEntry,
  size: "small" | "normal" | "large" | "art_crop" = "normal",
  face = 0,
): string | null {
  if (!entry.scryfall_id) return null;
  const base = `/catalog/image/${entry.scryfall_id}?size=${size}`;
  return face > 0 ? `${base}&face=${face}` : base;
}

/** Human copy for each completeness bucket. Single source for the page. */
export const COMPLETENESS_LABELS: Record<Completeness, string> = {
  full: "Complete",
  caveats: "Has caveats",
  unreviewed: "Unreviewed",
};

export const COMPLETENESS_BLURBS: Record<Completeness, string> = {
  full: "Every clause printed on the card is implemented.",
  caveats: "Works, but a printed clause is not modelled. The card says which.",
  unreviewed:
    "Automated, but nobody has checked it against the printed card yet. It may be complete or may have gaps.",
};
