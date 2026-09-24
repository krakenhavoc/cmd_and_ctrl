// roadmap.ts — the public engine roadmap: types, fetch, and the pure
// filtering and grouping the #/roadmap page runs over the registry.
//
// Mirrors server/internal/roadmap/roadmap.go (Roadmap / Counts /
// CardCounts / Entry / Example / PartialExample). Hand-maintained, like
// catalog.ts — update both sides in lockstep.
//
// The body carries card NAMES and caveat sentences only, never art or
// oracle text (ADR 0092 Decision 1), which is why the route and this
// page are public while the catalogue is not.

export type RoadmapKind = "keyword" | "mechanic" | "seam";
export type RoadmapStatus = "implemented" | "partial" | "missing";

export interface RoadmapExample {
  name: string;
  oracle_id: string;
}

export interface RoadmapPartialExample {
  name: string;
  caveat: string;
}

export interface RoadmapItem {
  slug: string;
  name: string;
  kind: RoadmapKind;
  status: RoadmapStatus;
  summary: string;
  /** What does not work yet. Absent on an implemented item. */
  missing?: string;
  /** Comprehensive Rules citations. */
  rules?: string[];
  /** Tracking issue number. */
  issue?: number;
  /** Decision record file name, e.g. "0092-public-roadmap-and-site-portal.md". */
  adr?: string;
  pin?: number;
  /** Cards this item alone blocks. */
  unblocks?: number;
  /** Fully automated cards that use it (at most three). */
  examples?: RoadmapExample[];
  /** Cards that work with a gap this item explains (at most three). */
  partial_examples?: RoadmapPartialExample[];
  /** Cards known to be waiting on it. */
  waiting?: string[];
}

export interface RoadmapCardCounts {
  total: number;
  full: number;
  caveats: number;
  unreviewed: number;
}

export interface RoadmapCounts {
  implemented: number;
  partial: number;
  missing: number;
  cards: RoadmapCardCounts;
}

export interface RoadmapResponse {
  counts: RoadmapCounts;
  /** "Up next" slugs, in order. */
  next: string[];
  items: RoadmapItem[];
}

/**
 * fetchRoadmap loads the published roadmap.
 *
 * Plain `fetch`, not `authFetch`: the route is public, and authFetch
 * reads any 401 as an expired session and clears it. Nobody should be
 * signed out by reading the roadmap.
 */
export async function fetchRoadmap(signal?: AbortSignal): Promise<RoadmapResponse> {
  const res = await fetch("/roadmap", {
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
  return (await res.json()) as RoadmapResponse;
}

export const KINDS: RoadmapKind[] = ["keyword", "mechanic", "seam"];
export const STATUSES: RoadmapStatus[] = ["implemented", "partial", "missing"];

export const KIND_LABELS: Record<RoadmapKind, string> = {
  keyword: "Keyword",
  mechanic: "Mechanic",
  seam: "Engine gap",
};

export const STATUS_LABELS: Record<RoadmapStatus, string> = {
  implemented: "Implemented",
  partial: "Partial",
  missing: "Missing",
};

export const STATUS_BLURBS: Record<RoadmapStatus, string> = {
  implemented: "The engine does all of it. Cards that use it resolve on their own.",
  partial: "Most of it works. The entry says which part you still resolve by hand.",
  missing: "Not built yet. Cards that need it still play, but you resolve that part by hand.",
};

export interface RoadmapFilter {
  /** Free text matched against name, summary, missing and card names. */
  query: string;
  /** Empty means any kind. Multiple selections are OR'd. */
  kinds: RoadmapKind[];
  /** Empty means any status. Multiple selections are OR'd. */
  statuses: RoadmapStatus[];
}

export const EMPTY_ROADMAP_FILTER: RoadmapFilter = { query: "", kinds: [], statuses: [] };

/**
 * matchesRoadmapQuery tests one item against free text:
 * case-insensitive substring over its name, slug, summary, missing
 * sentence, and every card name it lists — so searching a card finds
 * the entries that mention it.
 */
export function matchesRoadmapQuery(item: RoadmapItem, query: string): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  const haystack = [
    item.name,
    item.slug,
    item.summary,
    item.missing ?? "",
    ...(item.examples ?? []).map((e) => e.name),
    ...(item.partial_examples ?? []).flatMap((e) => [e.name, e.caveat]),
    ...(item.waiting ?? []),
  ]
    .join("\n")
    .toLowerCase();
  return haystack.includes(q);
}

/**
 * filterRoadmap applies the whole filter. Within a facet selections are
 * OR'd; across facets they are AND'd. Order is preserved.
 */
export function filterRoadmap(items: RoadmapItem[], filter: RoadmapFilter): RoadmapItem[] {
  return items.filter((it) => {
    if (filter.kinds.length > 0 && !filter.kinds.includes(it.kind)) return false;
    if (filter.statuses.length > 0 && !filter.statuses.includes(it.status)) return false;
    return matchesRoadmapQuery(it, filter.query);
  });
}

/**
 * groupByStatus splits items into the page's three sections, each
 * sorted by name (case-insensitive). Every status is present, possibly
 * empty, so the page can decide what to render.
 */
export function groupByStatus(items: RoadmapItem[]): Record<RoadmapStatus, RoadmapItem[]> {
  const out: Record<RoadmapStatus, RoadmapItem[]> = { implemented: [], partial: [], missing: [] };
  for (const it of items) {
    out[it.status]?.push(it);
  }
  for (const s of STATUSES) {
    out[s].sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: "base" }));
  }
  return out;
}

/**
 * nextItems resolves the "up next" slugs to items, in the server's
 * order, dropping any slug the list does not contain.
 */
export function nextItems(res: Pick<RoadmapResponse, "next" | "items">): RoadmapItem[] {
  const bySlug = new Map(res.items.map((it) => [it.slug, it]));
  return res.next.flatMap((slug) => {
    const it = bySlug.get(slug);
    return it ? [it] : [];
  });
}

/**
 * impactLine is the one-line "what this would unlock" note on an
 * unfinished item: the cards it alone blocks, then the cards recorded
 * as waiting on it. The two are kept apart because they are different
 * measurements (an audit's upper bound, and recorded skips), and the
 * server ranks "up next" by their sum. Empty when both are zero.
 */
export function impactLine(item: RoadmapItem): string {
  const parts: string[] = [];
  const u = item.unblocks ?? 0;
  const w = item.waiting?.length ?? 0;
  if (u > 0) parts.push(`unblocks ${u} ${u === 1 ? "card" : "cards"}`);
  if (w > 0) parts.push(`${w} waiting`);
  const line = parts.join(" · ");
  return line ? line[0].toUpperCase() + line.slice(1) : "";
}

const REPO = "https://github.com/krakenhavoc/cmd_and_ctrl";

/** issueURL links a tracking issue on the public repository. */
export function issueURL(n: number): string {
  return `${REPO}/issues/${n}`;
}

/** adrURL links a decision record on the develop branch. */
export function adrURL(file: string): string {
  return `${REPO}/blob/develop/docs/decisions/${encodeURIComponent(file)}`;
}

/** adrLabel turns "0092-public-roadmap.md" into "ADR 0092". */
export function adrLabel(file: string): string {
  const m = /^(\d{4})/.exec(file);
  return m ? `ADR ${m[1]}` : file.replace(/\.md$/, "");
}

/** catalogSearchHash is the catalogue deep link that pre-fills its search. */
export function catalogSearchHash(name: string): string {
  return `#/catalog?q=${encodeURIComponent(name)}`;
}
