// catalog.ts — the public card catalogue: types, fetch, and the
// filtering the page runs over a few hundred cards.
//
// Mirrors server/internal/catalog/catalog.go (Entry / Face / Counts /
// Response) and sets.go (SetSummary). Hand-maintained, like protocol.ts — update both sides in
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
  /**
   * Codes of the sets the card is printed in, newest set first; each is a
   * `CatalogSet.code`. Absent for a card with no counted printing (#4790).
   */
  sets?: string[];
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

/**
 * One set in the set picker (#4790): its unique cards (oracle ids, with
 * tokens, emblems, art-series cards, digital-only printings and basic lands
 * left out) and how many the engine plays. `full + caveats + unreviewed` is
 * the supported part; `unsupported` is the rest, so the four add up to
 * `total`. Sorted newest first by the server.
 */
export interface CatalogSet {
  code: string;
  name: string;
  /** "YYYY-MM-DD". */
  released?: string;
  total: number;
  full: number;
  caveats: number;
  unreviewed: number;
  unsupported: number;
}

export interface CatalogResponse {
  counts: CatalogCounts;
  cards: CatalogEntry[];
  /** Absent when the server has no card data loaded. */
  sets?: CatalogSet[];
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
  /** A set code, or "" for any set. One set at a time (#4790). */
  set: string;
}

export const EMPTY_FILTER: CatalogFilter = {
  query: "",
  colors: [],
  types: [],
  completeness: [],
  set: "",
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
    if (filter.set && !(e.sets ?? []).includes(filter.set)) return false;
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

/** Per-value counts for the sidebar's Completeness and Type facets. */
export interface FacetCounts {
  completeness: Record<Completeness, number>;
  types: Record<TypeFilter, number>;
}

/**
 * facetCounts counts the cards each sidebar value would show.
 *
 * Each facet is counted with every OTHER filter applied and its own
 * cleared, so a count says "tick this and you get N" — the number a
 * person wants before they click, and one that never hides a value
 * they have already ticked. Computed from the catalogue the page has
 * loaded; there is no server round trip.
 */
export function facetCounts(entries: CatalogEntry[], filter: CatalogFilter): FacetCounts {
  const completeness: Record<Completeness, number> = { full: 0, caveats: 0, unreviewed: 0 };
  for (const e of filterCatalog(entries, { ...filter, completeness: [] })) {
    completeness[e.completeness]++;
  }
  const types = Object.fromEntries(TYPE_FILTERS.map((t) => [t, 0])) as Record<TypeFilter, number>;
  for (const e of filterCatalog(entries, { ...filter, types: [] })) {
    for (const t of TYPE_FILTERS) if ((e.types ?? []).includes(t)) types[t]++;
  }
  return { completeness, types };
}

/** One removable chip over the grid: a single active filter value. */
export type FilterChip =
  | { kind: "query"; value: string; label: string }
  | { kind: "colors"; value: ColorFilter; label: string }
  | { kind: "types"; value: TypeFilter; label: string }
  | { kind: "completeness"; value: Completeness; label: string }
  | { kind: "set"; value: string; label: string };

/**
 * activeChips lists one chip per active filter value, in sidebar order.
 * `sets` names the selected set; without it the chip reads the set's code.
 */
export function activeChips(filter: CatalogFilter, sets: CatalogSet[] = []): FilterChip[] {
  const chips: FilterChip[] = [];
  if (filter.set) {
    chips.push({
      kind: "set",
      value: filter.set,
      label: findSet(sets, filter.set)?.name ?? filter.set.toUpperCase(),
    });
  }
  const q = filter.query.trim();
  if (q) chips.push({ kind: "query", value: q, label: `“${q}”` });
  for (const b of filter.completeness) {
    chips.push({ kind: "completeness", value: b, label: COMPLETENESS_LABELS[b] });
  }
  for (const c of filter.colors) chips.push({ kind: "colors", value: c, label: COLOR_LABELS[c] });
  for (const t of filter.types) chips.push({ kind: "types", value: t, label: t });
  return chips;
}

/** removeChip returns the filter without the chip's value. */
export function removeChip(filter: CatalogFilter, chip: FilterChip): CatalogFilter {
  switch (chip.kind) {
    case "query":
      return { ...filter, query: "" };
    case "colors":
      return { ...filter, colors: filter.colors.filter((v) => v !== chip.value) };
    case "types":
      return { ...filter, types: filter.types.filter((v) => v !== chip.value) };
    case "completeness":
      return { ...filter, completeness: filter.completeness.filter((v) => v !== chip.value) };
    case "set":
      return { ...filter, set: "" };
  }
}

/** findSet looks a set up by its code. */
export function findSet(sets: CatalogSet[], code: string): CatalogSet | undefined {
  return sets.find((s) => s.code === code);
}

/** supportedCount is how many of a set's cards the engine plays at all. */
export function supportedCount(s: CatalogSet): number {
  return s.full + s.caveats + s.unreviewed;
}

/**
 * setOptionLabel is a set's line in the picker: the name, then how many of
 * its cards are supported out of how many it has ("Marvel's Spider-Man
 * (120/188)"). The picker's options carry it as their accessible name.
 */
export function setOptionLabel(s: CatalogSet): string {
  return `${s.name} (${supportedCount(s)}/${s.total})`;
}

/**
 * searchSets is the set picker's filter (#4790): the sets whose name
 * contains the query, or whose code starts with it, best match first.
 * Players name a set by its code as often as by its name ("hml",
 * "MH3", "tmt"), so a code that equals the query outright ranks first,
 * then the names that start with it, then the other names that contain
 * it, then the codes that only start with it. Case-insensitive. Within a
 * group the sets keep the order they came in, which from the server is
 * newest first; an empty query returns every set in that order.
 */
export function searchSets(sets: CatalogSet[], query: string): CatalogSet[] {
  const q = query.trim().toLowerCase();
  if (!q) return sets.slice();
  const exactCode: CatalogSet[] = [];
  const namePrefix: CatalogSet[] = [];
  const nameContains: CatalogSet[] = [];
  const codePrefix: CatalogSet[] = [];
  for (const s of sets) {
    const code = s.code.toLowerCase();
    const name = s.name.toLowerCase();
    if (code === q) exactCode.push(s);
    else if (name.startsWith(q)) namePrefix.push(s);
    else if (name.includes(q)) nameContains.push(s);
    else if (code.startsWith(q)) codePrefix.push(s);
  }
  return [...exactCode, ...namePrefix, ...nameContains, ...codePrefix];
}

/** setYear is the year a set came out, or "" when its date is unknown. */
export function setYear(s: CatalogSet): string {
  return /^\d{4}/.test(s.released ?? "") ? s.released!.slice(0, 4) : "";
}

/**
 * setSummary is the one-line answer for the selected set: how many of its
 * cards the engine plays, what the supported ones are like, and how many
 * it does not play yet. Parts that are zero are left out.
 *
 *   Marvel's Spider-Man: 120 of 188 supported — 118 complete, 2 with caveats; 68 not yet supported
 */
export function setSummary(s: CatalogSet): string {
  const supported = supportedCount(s);
  if (s.total === 0) return `${s.name}: no cards to support.`;
  if (supported === 0) {
    return s.total === 1
      ? `${s.name}: its one card is not supported yet.`
      : `${s.name}: none of its ${s.total} cards supported yet.`;
  }
  const parts: string[] = [];
  if (s.full > 0) parts.push(`${s.full} complete`);
  if (s.caveats > 0) parts.push(`${s.caveats} with caveats`);
  if (s.unreviewed > 0) parts.push(`${s.unreviewed} unreviewed`);
  const head = `${s.name}: ${supported} of ${s.total} supported — ${parts.join(", ")}`;
  return s.unsupported > 0 ? `${head}; ${s.unsupported} not yet supported` : `${head}.`;
}

export const COMPLETENESS_BLURBS: Record<Completeness, string> = {
  full: "Every clause printed on the card is implemented.",
  caveats: "Works, but a printed clause is not modelled. The card says which.",
  unreviewed:
    "Automated, but nobody has checked it against the printed card yet. It may be complete or may have gaps.",
};
