import type { GameMeta, SeatInfo } from "./api";
import { myGameStatus, relativeDay, type MyGame } from "./myGames";

// tablesPage.ts is the pure half of the Tables page (ADR 0158 §1): which
// tables are Now, the History rows and their filters, the status chip,
// the card's line of context and the seat's second line. No DOM and no
// fetch, so all of it is unit-tested (tablesPage.test.ts); Lobby.svelte
// only renders the answers.

// HISTORY_HASH opens Tables scrolled to History. #/my-games redirects
// here (§1 item 5), so a bookmark to the old page still lands.
export const HISTORY_HASH = "#/lobby?show=history";
// The id of the History section, which HISTORY_HASH scrolls to.
export const HISTORY_ID = "history";

// The History list's length (§1 item 3): the first six rows, then
// "Show N more", which adds twenty at a time.
export const HISTORY_FIRST = 6;
export const HISTORY_STEP = 20;

// nowTables are the tables in the lobby or in play, from the tables the
// page lists (signedInHome.myTables). Ended and archived tables are
// History's.
export function nowTables(listed: GameMeta[]): GameMeta[] {
  return listed.filter((g) => !g.archived_at && (g.state === "lobby" || g.state === "active"));
}

// statusChip is the card head's chip: Lobby, or In play · turn N. A game
// in its opening roll has no turn yet (GET /games omits it).
export function statusChip(g: GameMeta): { label: string; live: boolean } {
  if (g.state === "active") {
    return { label: g.turn ? `In play · turn ${g.turn}` : "In play", live: true };
  }
  if (g.state === "ended") return { label: "Ended", live: false };
  return { label: "Lobby", live: false };
}

// seatName is a seat's shown name: Discord's display name, else the
// name it joined with.
export function seatName(p: SeatInfo): string {
  return p.display_name || p.name;
}

// timeAgo is a short relative time: "just now", "5m ago", "2h ago",
// "3d ago".
export function timeAgo(iso: string, now: number = Date.now()): string {
  const ms = now - Date.parse(iso);
  if (!Number.isFinite(ms) || ms < 0) return "just now";
  const m = Math.floor(ms / 60_000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

// contextLine is the card head's line of context: who hosts, your seat
// when someone else does, and in the lobby how long ago it was made.
// `mySeat` is the seat this person holds at the table, through this
// session or another (GET /me/games), or null.
export function contextLine(
  g: GameMeta,
  me: { seat: number | null },
  now: number = Date.now(),
): string {
  const parts: string[] = [];
  const host = g.players.find((p) => p.is_host);
  const iHost = host !== undefined && me.seat !== null && host.seat === me.seat;
  if (iHost) parts.push("you host");
  else if (host) parts.push(`${seatName(host)} hosts`);
  else if (g.is_creator) parts.push("you created it");
  if (me.seat !== null && !iHost) parts.push(`your seat ${me.seat + 1}`);
  if (g.state === "lobby") parts.push(`created ${timeAgo(g.created_at, now)}`);
  return parts.join(" · ");
}

// seatLine is a taken seat's second line: the deck in the lobby, the
// deck and life total in play. A bot names its tier and deck. power is
// the seated deck's estimated power level, 0 when unrated (#3671), kept
// out of text because the line is cut off on a narrow seat: it renders
// as a DeckPowerBadge ahead of the text.
export function seatLine(
  g: GameMeta,
  p: SeatInfo,
  botDeckName: (id: string | undefined) => string,
): { text: string; pending: boolean; power: number } {
  const life = g.state === "active" && typeof p.life === "number" ? `${p.life} life` : "";
  let deck: string;
  if (p.is_bot) {
    const tier = p.bot_own_key ? `${p.bot_tier || "bot"} (own key)` : p.bot_tier || "bot";
    deck = [tier, botDeckName(p.bot_deck) || p.deck_name || ""].filter(Boolean).join(" · ");
  } else if (p.deck_uploaded) {
    deck = p.deck_name || "deck ready";
  } else if (g.state === "lobby") {
    return { text: "deck pending", pending: true, power: 0 };
  } else {
    deck = "";
  }
  const power = !p.is_bot && p.deck_uploaded ? (p.deck_power ?? 0) : 0;
  return { text: [deck, life].filter(Boolean).join(" · "), pending: false, power };
}

// --- History (§1 item 3) ---------------------------------------------

export type HistoryFilter = "all" | "won" | "friends" | "archived";

export const HISTORY_FILTERS: ReadonlyArray<{ id: HistoryFilter; label: string }> = [
  { id: "all", label: "All" },
  { id: "won", label: "Won" },
  { id: "friends", label: "With friends" },
  { id: "archived", label: "Archived" },
];

export type ResultTone = "won" | "ended" | "archived" | "open";

export interface HistoryRow {
  id: string;
  name: string;
  state: GameMeta["state"];
  result: { label: string; tone: ResultTone };
  // Who else was there, in seat order, or "" for nobody.
  others: string;
  // When, in Unix milliseconds: the latest thing that happened to it.
  at: number;
  won: boolean;
  withFriends: boolean;
  archived: boolean;
  // Whether this person held a seat (a GET /me/games entry): their
  // replay of a finished game is theirs to download.
  seated: boolean;
  // The path that trades this session for the seat (POST
  // /me/games/{id}/session), while the table is still open.
  rejoin?: string;
}

function capitalise(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function fromMine(g: MyGame): HistoryRow {
  const status = myGameStatus(g);
  const others = [...(g.others ?? [])].sort((a, b) => a.seat - b.seat);
  const tone: ResultTone =
    status.tone === "won"
      ? "won"
      : status.tone === "archived"
        ? "archived"
        : status.tone === "ended"
          ? "ended"
          : "open";
  return {
    id: g.id,
    name: g.name,
    state: g.state,
    result: { label: capitalise(status.label), tone },
    others: others
      .map((o) => o.name)
      .filter((n) => n && n.trim() !== "")
      .join(", "),
    at: g.archived_at ?? g.ended_at ?? g.started_at ?? g.created_at,
    won: status.tone === "won",
    withFriends: others.some((o) => !o.bot),
    archived: g.archived_at != null,
    seated: true,
    rejoin: g.rejoin,
  };
}

// fromListed is a History row for a table this person sees in GET
// /games but not in GET /me/games: a table they created and never sat
// at, or any table for an admin. The list carries no winner, so its
// result is Ended or Archived.
function fromListed(g: GameMeta): HistoryRow {
  const archived = Boolean(g.archived_at);
  const at = Date.parse(g.archived_at ?? g.created_at);
  return {
    id: g.id,
    name: g.name,
    state: g.state,
    result: archived
      ? { label: "Archived", tone: "archived" }
      : g.state === "ended"
        ? { label: "Ended", tone: "ended" }
        : { label: g.state === "active" ? "In play" : "Lobby", tone: "open" },
    others: [...g.players]
      .sort((a, b) => a.seat - b.seat)
      .map(seatName)
      .join(", "),
    at: Number.isFinite(at) ? at : 0,
    won: false,
    withFriends: g.players.filter((p) => !p.is_bot).length >= 2,
    archived,
    seated: false,
  };
}

// historyRows is every table this person has sat at that ended or was
// archived (GET /me/games minus the tables in Now), plus the ended and
// archived tables the page already lists for them (an admin's whole
// room, a creator's own tables). One row per table, a seat's entry
// first, most recent first.
export function historyRows(
  mine: MyGame[],
  listed: GameMeta[],
  nowIDs: ReadonlySet<string>,
): HistoryRow[] {
  const rows = new Map<string, HistoryRow>();
  for (const g of mine) {
    if (nowIDs.has(g.id) || rows.has(g.id)) continue;
    rows.set(g.id, fromMine(g));
  }
  for (const g of listed) {
    if (nowIDs.has(g.id) || rows.has(g.id)) continue;
    if (g.state !== "ended" && !g.archived_at) continue;
    rows.set(g.id, fromListed(g));
  }
  return [...rows.values()].sort((a, b) => b.at - a.at || a.id.localeCompare(b.id));
}

export function filterHistory(rows: HistoryRow[], f: HistoryFilter): HistoryRow[] {
  switch (f) {
    case "won":
      return rows.filter((r) => r.won);
    case "friends":
      return rows.filter((r) => r.withFriends);
    case "archived":
      return rows.filter((r) => r.archived);
    default:
      return rows;
  }
}

// moreCount is how many rows "Show N more" adds: twenty, or what is
// left.
export function moreCount(total: number, shown: number): number {
  return Math.max(0, Math.min(HISTORY_STEP, total - shown));
}

// whenLabel is a History row's date (myGames.relativeDay).
export function whenLabel(r: HistoryRow, now: number = Date.now()): string {
  return relativeDay(r.at, now);
}
