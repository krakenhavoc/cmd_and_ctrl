import type { BotDeckInfo, BotTierInfo, GameMeta, SeatInfo } from "./api";
import type { DeckArt } from "./deckGrid";
import { sameName, type DeckPreview, type PickerCommit } from "./deckPicker";
import { L } from "./labels";
import { myGameStatus, relativeDay, type MyGame } from "./myGames";
import { orderedColors } from "./prebuiltDecks";

// tablesPage.ts is the pure half of the Tables page (ADR 0158 §1): which
// tables are Now, the History rows and their filters, the status chip,
// the card's line of context and the seat's second line, and for a table
// in the lobby its seat tiles, readiness line and primary (ADR 0212). No
// DOM and no fetch, so all of it is unit-tested (tablesPage.test.ts);
// Lobby.svelte only renders the answers.

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

// --- A table in the lobby: a character select (ADR 0212) -------------

// A table seats four; there is no table-size setting (ADR 0212 owner
// decision 5), so a lobby card always draws four tiles.
export const TABLE_SEATS = 4;

// lobbySlots is a lobby table's four seats by seat number, null where a
// seat is open. The lobby renumbers seats when one leaves, so the open
// ones are the last; a seat number out of range is kept, at the end,
// rather than dropped.
export function lobbySlots(g: GameMeta): (SeatInfo | null)[] {
  const slots: (SeatInfo | null)[] = Array.from({ length: TABLE_SEATS }, () => null);
  const extra: SeatInfo[] = [];
  for (const p of [...g.players].sort((a, b) => a.seat - b.seat)) {
    if (p.seat >= 0 && p.seat < TABLE_SEATS && slots[p.seat] === null) slots[p.seat] = p;
    else extra.push(p);
  }
  return [...slots, ...extra];
}

// seatReady: a seat with its deck in, which is what Start waits for
// (owner decision 4: no Ready toggle). A bot arrives with its deck.
export function seatReady(p: SeatInfo): boolean {
  return p.is_bot === true || p.deck_uploaded;
}

// SeatTileText is what a taken seat's tile says under its name (§1):
// the commander, a quieter second line only when it adds something,
// the colour identity as WUBRG letters, and whether it is ready.
export interface SeatTileText {
  what: string;
  second: string;
  colors: string[];
  ready: boolean;
}

// seatTileText reads a seat's tile. Until the seats carry their
// commanders (phase B2), a person's tile names the deck once; a bot's
// names its pre-built deck's commander, then "Pre-built · <deck>", with
// the deck's colours from GET /bot/options.
export function seatTileText(p: SeatInfo, botDeck: BotDeckInfo | undefined): SeatTileText {
  if (p.is_bot) {
    const deck = botDeck?.name || p.deck_name || "";
    return {
      what: botDeck?.commander || deck || "Pre-built deck",
      second: deck ? `Pre-built · ${deck}` : "",
      colors: orderedColors(botDeck?.colors),
      ready: true,
    };
  }
  if (!p.deck_uploaded) return { what: "No deck yet", second: "", colors: [], ready: false };
  return { what: p.deck_name || "Deck ready", second: "", colors: [], ready: true };
}

// SeatPreview is your own tile while the deck picker is open (§1).
export interface SeatPreview {
  text: SeatTileText;
  art: DeckArt[];
  // Marked "Preview" and "Not seated yet".
  preview: boolean;
}

// seatPreview: your tile shows the deck selected in the picker, art and
// all, marked Preview and not ready, since nothing is sent until you sit
// down. The deck already in your seat (the same name: the last deck you
// seated is preselected) is drawn ready, with its art. null with nothing
// selected, and your tile says what it says without the picker.
export function seatPreview(p: SeatInfo, pv: DeckPreview | null): SeatPreview | null {
  if (!pv) return null;
  const seated = p.deck_uploaded && sameName(p.deck_name ?? "", pv.name);
  return { text: { ...pv.text, ready: seated }, art: pv.art, preview: !seated };
}

// botChipText is a bot seat's chip: its tier's name, and "(own key)"
// when it plays on the model key of the player who added it (ADR 0182).
export function botChipText(p: SeatInfo, tiers: readonly BotTierInfo[]): string {
  const tier = tiers.find((t) => t.tier === p.bot_tier)?.label || p.bot_tier || "Bot";
  return p.bot_own_key ? `${tier} (own key)` : tier;
}

const plural = (n: number, word: string): string => `${n} ${word}${n === 1 ? "" : "s"}`;

// A readiness segment per seat, in seat order (§2): filled when ready,
// outlined while choosing, dashed when open.
export type SegmentState = "ready" | "choosing" | "open";

export interface Readiness {
  ready: number;
  // Seated players, bots included: open seats do not block Start, so
  // the line counts against these, not four.
  seated: number;
  segments: { seat: number; state: SegmentState }[];
  // The sentence's two halves: the open seats, then what the table is
  // waiting on. The first one present starts with a capital.
  lead: string;
  status: string;
}

// readiness is a lobby table's readiness line (§2): "READY 2/3", the
// segments, and "1 seat open · waiting on your deck". `bots` says
// whether this person can add a bot, for the advice at fewer than two.
export function readiness(g: GameMeta, me: SeatInfo | null, opts: { bots: boolean }): Readiness {
  const slots = lobbySlots(g);
  const segments = slots.map((p, seat) => ({
    seat: p?.seat ?? seat,
    state: (p === null ? "open" : seatReady(p) ? "ready" : "choosing") as SegmentState,
  }));
  const seated = g.players.length;
  const ready = g.players.filter(seatReady).length;
  const open = Math.max(0, TABLE_SEATS - seated);
  const choosing = g.players.filter((p) => !seatReady(p));
  let status: string;
  if (me && !seatReady(me)) status = "waiting on your deck";
  else if (choosing.length === 1) status = `waiting on ${seatName(choosing[0])}'s deck`;
  else if (choosing.length > 1) status = `waiting on ${choosing.length} decks`;
  else if (seated < 2) status = opts.bots ? "invite someone or add a bot" : "invite someone";
  else status = "every deck is in";
  const lead = open > 0 ? plural(open, "seat") + " open" : "";
  return { ready, seated, segments, lead, status: lead ? status : capitalise(status) };
}

// What the card's one primary does (§2). The page wires each kind to its
// action: open (the board, or reclaiming a seat held elsewhere), take a
// seat, choose a deck, sit down with the deck the picker holds, add the
// bot the bot picker holds, start, or watch; a wait is disabled.
export type LobbyPrimaryKind =
  | "open"
  | "take"
  | "choose"
  | "sit"
  | "bot"
  | "wait"
  | "start"
  | "watch";

export interface LobbyPrimary {
  kind: LobbyPrimaryKind;
  label: string;
  disabled: boolean;
  // The tooltip: what it does, or why it is disabled (DESIGN.md §6).
  title: string;
}

export interface LobbyViewer {
  // The seat this session holds at the table, or null.
  me: SeatInfo | null;
  // This person holds a seat here through another session.
  heldElsewhere: boolean;
  // The creator, not seated, holding the invite to sit down with.
  takesSeat: boolean;
  // This person may start the table: a seated host or creator, or the
  // admin, as the server's gate.
  mayStart: boolean;
  // A spectator session of this table (owner decision 6).
  spectator: boolean;
  // The deck picker is open, and this is what committing it says (§2
  // rule 4): Sit down with this deck, or Upload / Import deck. null when
  // the open tab holds no deck.
  picking?: PickerCommit | null;
  // The bot picker is open (owner decision 8), and this adds its bot.
  addingBot?: PickerCommit | null;
}

// lobbyPrimary picks a lobby table's one primary: the first rule of
// ADR 0212 §2 that applies wins. The pickers under the seats hold no
// commit of their own: while the deck picker is open with a deck, the
// primary seats it (rule 4), and while the bot picker is open it adds
// the bot (owner decision 8), so the card keeps one primary. An empty
// Paste a list box takes the primary only from a seat with no deck yet,
// where Upload deck, disabled, says what to do; a seat that has one
// keeps waiting or starting.
export function lobbyPrimary(g: GameMeta, v: LobbyViewer): LobbyPrimary {
  const open = (title: string): LobbyPrimary => ({
    kind: "open",
    label: L.openTable,
    disabled: false,
    title,
  });
  const wait = (who: string, title: string): LobbyPrimary => ({
    kind: "wait",
    label: L.waitingFor(who),
    disabled: true,
    title,
  });
  const choosing = g.players.filter((p) => !seatReady(p));
  const start = (): LobbyPrimary => {
    const n = g.players.length;
    return {
      kind: "start",
      label: L.startGame(n),
      disabled: false,
      title:
        n < TABLE_SEATS
          ? `start the game with ${n} players; the open ${n === TABLE_SEATS - 1 ? "seat closes" : "seats close"}`
          : "start the game",
    };
  };
  const commit = (kind: "sit" | "bot", c: PickerCommit): LobbyPrimary => ({
    kind,
    label: c.label,
    disabled: c.disabled,
    title: c.title,
  });
  if (v.heldElsewhere) return open("open your seat at this table");
  if (v.addingBot) return commit("bot", v.addingBot);
  if (v.takesSeat) {
    return { kind: "take", label: "Take a seat", disabled: false, title: "sit down at your table" };
  }
  if (v.me) {
    if (v.picking && !(v.picking.empty && seatReady(v.me))) return commit("sit", v.picking);
    if (!seatReady(v.me)) {
      return {
        kind: "choose",
        label: L.chooseYourDeck,
        disabled: false,
        title: "pick a deck for your seat",
      };
    }
    if (choosing.length > 0) {
      return wait(
        plural(choosing.length, "player"),
        choosing.length === 1
          ? `waiting on ${seatName(choosing[0])}'s deck`
          : `waiting on ${choosing.length} decks`,
      );
    }
    if (g.players.length < 2) {
      return wait("players", "a game needs two players: invite someone or add a bot");
    }
    if (v.mayStart) return start();
    const host = g.players.find((p) => p.is_host);
    return wait(
      host ? `${seatName(host)} to start` : "the host to start",
      "only the table's host can start the game",
    );
  }
  if (v.mayStart && choosing.length === 0 && g.players.length >= 2) return start();
  if (v.spectator) {
    return { kind: "watch", label: "Watch", disabled: false, title: "watch this table" };
  }
  return open("open this table");
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
