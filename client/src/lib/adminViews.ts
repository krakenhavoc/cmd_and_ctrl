// adminViews.ts — the pure half of ADR 0124 §7, the admin views:
// Live now, Games, one table, Accounts and one account.
//
// What lives here: the routes' parsing and building (a filter built
// into a hash parses back to itself, and a value the server does not
// know is dropped, never sent), the response shapes of the five
// read-only routes, the labels a row needs (a seat, an outcome, a
// relative time), the client-side sort and search of the accounts list,
// the confirmation copy of the linked actions, and Live now's poll
// schedule. No DOM and no fetch: the fetchers are in api.ts and the
// pages in routes/Admin.svelte and lib/components/admin/.
//
// The server is the gate (requireAdmin on every route). Nothing here
// decides who may see a page; lib/admin.ts's isAdmin only decides
// whether the page asks at all.

// --- the routes -------------------------------------------------------

export const GAME_STATES = ["lobby", "active", "ended"] as const;
export type GameState = (typeof GAME_STATES)[number];

export const ARCHIVED_VALUES = ["true", "false"] as const;
export type ArchivedFilter = (typeof ARCHIVED_VALUES)[number];

// `practice` defaults to `exclude` on the server (ADR 0124 §2), so the
// unfiltered list counts what the Grafana tiles count.
export const PRACTICE_VALUES = ["exclude", "include", "only"] as const;
export type PracticeFilter = (typeof PRACTICE_VALUES)[number];

export const PLAYED_WINDOWS = ["1d", "7d", "30d"] as const;
export type PlayedWindow = (typeof PLAYED_WINDOWS)[number];

// `sort` is client-only (§7): the server never sees it.
export const ACCOUNT_SORTS = ["last_played", "first_seen", "name", "games"] as const;
export type AccountSort = (typeof ACCOUNT_SORTS)[number];

/** The Games view's filters: §2's parameters, as the hash carries them. */
export interface GamesFilter {
  state?: GameState;
  archived?: ArchivedFilter;
  practice?: PracticeFilter;
  /** an account id: the tables it holds a seat at */
  user?: string;
  /** an opaque `next_cursor` from the page before */
  cursor?: string;
}

/** The Accounts view's filters. */
export interface AccountsFilter {
  played?: PlayedWindow;
  sort?: AccountSort;
}

/** Which admin view a hash names (`#/admin/<view>[/<id>][?<filter>]`). */
export type AdminView =
  | { view: "live" }
  | { view: "games"; filter: GamesFilter }
  | { view: "game"; id: string }
  | { view: "accounts"; filter: AccountsFilter }
  | { view: "account"; id: string };

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/** isUUID reports whether s is a UUID, the shape of every account and game id. */
export function isUUID(s: string | null | undefined): s is string {
  return typeof s === "string" && UUID.test(s);
}

// A cursor is the server's, and opaque. Anything printable and short is
// passed back as it came; anything else is dropped.
const CURSOR = /^[A-Za-z0-9_\-=.~]{1,512}$/;

function oneOf<T extends string>(values: readonly T[], v: string | null): T | undefined {
  return v !== null && (values as readonly string[]).includes(v) ? (v as T) : undefined;
}

/** gamesFilterFrom keeps the known values of a Games hash's query. */
export function gamesFilterFrom(params: URLSearchParams): GamesFilter {
  const f: GamesFilter = {};
  const state = oneOf(GAME_STATES, params.get("state"));
  if (state) f.state = state;
  const archived = oneOf(ARCHIVED_VALUES, params.get("archived"));
  if (archived) f.archived = archived;
  const practice = oneOf(PRACTICE_VALUES, params.get("practice"));
  if (practice) f.practice = practice;
  const user = params.get("user");
  if (isUUID(user)) f.user = user.toLowerCase();
  const cursor = params.get("cursor");
  if (cursor && CURSOR.test(cursor)) f.cursor = cursor;
  return f;
}

/** accountsFilterFrom keeps the known values of an Accounts hash's query. */
export function accountsFilterFrom(params: URLSearchParams): AccountsFilter {
  const f: AccountsFilter = {};
  const played = oneOf(PLAYED_WINDOWS, params.get("played"));
  if (played) f.played = played;
  const sort = oneOf(ACCOUNT_SORTS, params.get("sort"));
  if (sort) f.sort = sort;
  return f;
}

/**
 * parseAdminView reads the path after `#/admin` (`rest`, already split
 * on "/") and its query. It returns null for `#/admin` alone, which
 * stays the shared token's form, and for any path it does not know.
 */
export function parseAdminView(rest: string[], params: URLSearchParams): AdminView | null {
  const [view, id, ...extra] = rest;
  if (extra.length > 0) return null;
  switch (view) {
    case "live":
      return id === undefined ? { view: "live" } : null;
    case "games":
      if (id !== undefined) return id ? { view: "game", id } : null;
      return { view: "games", filter: gamesFilterFrom(params) };
    case "accounts":
      if (id !== undefined) return id ? { view: "account", id } : null;
      return { view: "accounts", filter: accountsFilterFrom(params) };
    default:
      return null;
  }
}

function gamesParams(f: GamesFilter): URLSearchParams {
  const p = new URLSearchParams();
  const clean = gamesFilterFrom(
    new URLSearchParams(
      Object.entries(f).filter((e): e is [string, string] => typeof e[1] === "string"),
    ),
  );
  if (clean.state) p.set("state", clean.state);
  if (clean.archived) p.set("archived", clean.archived);
  if (clean.practice) p.set("practice", clean.practice);
  if (clean.user) p.set("user", clean.user);
  if (clean.cursor) p.set("cursor", clean.cursor);
  return p;
}

function withQuery(path: string, p: URLSearchParams): string {
  const q = p.toString();
  return q ? `${path}?${q}` : path;
}

/** gamesHash is the Games view's hash for f. Unknown values are left out. */
export function gamesHash(f: GamesFilter = {}): string {
  return withQuery("#/admin/games", gamesParams(f));
}

/** gamesQuery is the server's URL for f: `GET /admin/games?…`. */
export function gamesQuery(f: GamesFilter = {}): string {
  return withQuery("/admin/games", gamesParams(f));
}

/** accountsHash is the Accounts view's hash for f, `sort` included. */
export function accountsHash(f: AccountsFilter = {}): string {
  const clean = accountsFilterFrom(
    new URLSearchParams(
      Object.entries(f).filter((e): e is [string, string] => typeof e[1] === "string"),
    ),
  );
  const p = new URLSearchParams();
  if (clean.played) p.set("played", clean.played);
  if (clean.sort) p.set("sort", clean.sort);
  return withQuery("#/admin/accounts", p);
}

/** accountsQuery is the server's URL for f: `sort` is never sent. */
export function accountsQuery(f: AccountsFilter = {}): string {
  const played = oneOf(PLAYED_WINDOWS, f.played ?? null);
  return played ? `/admin/users?played=${played}` : "/admin/users";
}

/** gameHash and accountHash are the detail views' hashes. */
export function gameHash(id: string): string {
  return `#/admin/games/${encodeURIComponent(id)}`;
}
export function accountHash(id: string): string {
  return `#/admin/accounts/${encodeURIComponent(id)}`;
}

/** adminViewHash is the canonical hash of a parsed view. */
export function adminViewHash(v: AdminView): string {
  switch (v.view) {
    case "live":
      return LIVE_HASH;
    case "games":
      return gamesHash(v.filter);
    case "game":
      return gameHash(v.id);
    case "accounts":
      return accountsHash(v.filter);
    case "account":
      return accountHash(v.id);
  }
}

export const LIVE_HASH = "#/admin/live";

/** The tab a view belongs to in the `navigation "admin views"` strip. */
export type AdminTab = "live" | "games" | "accounts";

export function tabOf(v: AdminView): AdminTab {
  if (v.view === "game") return "games";
  if (v.view === "account") return "accounts";
  return v.view;
}

// --- the responses (docs/lobby.md, "Admin views" and "Live now") ------

/** A person's account as a row names it. */
export interface AdminAccountRef {
  id: string;
  name?: string;
  avatar_url?: string;
}

/** A seat (ADR 0124 §3.3), as every view serves it. */
export interface AdminSeat {
  seat: number;
  kind: string; // "human" | "bot" | "agent"
  account?: AdminAccountRef;
  guest_name?: string;
  discord_pending?: boolean;
  bot_tier?: string;
  agent_client?: string;
  deck_name?: string;
  host: boolean;
  /** live sockets bound to the seat; a loaded table only, absent never means zero */
  connected?: number;
  /** the earliest of those sockets' connection times (Live now) */
  since?: number;
}

/** A table row (§3.3). */
export interface AdminGame {
  id: string;
  name: string;
  state: string;
  created_at?: number;
  started_at?: number;
  ended_at?: number;
  archived_at?: number;
  outcome?: string; // "win" | "draw" | "closed"
  winner_seat?: number;
  creator?: { id: string; name: string };
  practice: boolean;
  loaded: boolean;
  seats: AdminSeat[];
  spectators_connected?: number;
  /** on an account's games: the seat that account holds */
  their_seat?: number;
}

/** One live socket at a table (`GET /admin/games/{id}`'s `connections`). */
export interface AdminConnection {
  kind: string; // "seat" | "spectator" | "admin"
  seat?: number;
  account?: AdminAccountRef;
  since?: number;
}

export interface AdminGamesResponse {
  generated_at: number;
  games: AdminGame[];
  next_cursor?: string;
}

export interface AdminGameResponse extends AdminGame {
  generated_at: number;
  connections?: AdminConnection[];
}

/** An account row (§3.1). */
export interface AdminAccount {
  id: string;
  name: string;
  avatar_url?: string;
  first_seen_at?: number;
  last_sign_in_at?: number;
  games_played: number;
  last_played_at?: number;
  playing_now: boolean;
}

export interface AdminAccountsResponse {
  generated_at: number;
  accounts: AdminAccount[];
  truncated: boolean;
}

export interface AdminSignIn {
  last_sign_in_at?: number;
  discord_linked_at?: number;
  sessions_invalid_before?: number;
  revoke_path: string;
}

export interface AdminDeck {
  id: string;
  name: string;
  format: string;
  source_url?: string;
  commanders: string[];
  card_count: number;
  created_at?: number;
  updated_at?: number;
}

export interface AdminDeckRequest {
  deck_key: string;
  asked_at?: number;
  issue_number?: number;
  issue_url?: string;
}

export interface AdminAccountResponse {
  generated_at: number;
  account: AdminAccount;
  sign_in: AdminSignIn;
  /** The account's playmat path, absent for none (ADR 0124 amendment). */
  playmat_url?: string;
  games: AdminGame[];
  games_truncated: boolean;
  decks: AdminDeck[];
  deck_requests: AdminDeckRequest[];
  deck_requests_truncated: boolean;
}

export interface LiveSpectator {
  account?: AdminAccountRef;
  since: number;
}

export interface LiveAdmin {
  account?: AdminAccountRef;
  as_seat?: number;
  since: number;
}

export interface LiveTable {
  id: string;
  name: string;
  state: string;
  practice: boolean;
  archived: boolean;
  seats: AdminSeat[];
  spectators: LiveSpectator[];
  admins: LiveAdmin[];
}

export interface LiveTotals {
  players_connected: number;
  spectators: number;
  bot_seats: number;
  practice_tables: number;
  admin_views: number;
}

export interface LiveNowResponse {
  generated_at: number;
  tables: LiveTable[];
  unbound_sockets: number;
  totals: LiveTotals;
}

// --- labels -----------------------------------------------------------

/** seatNumber is a seat as people count them: seat 0 is "seat 1". */
export function seatNumber(seat: number): string {
  return `seat ${seat + 1}`;
}

/**
 * seatName is who holds a seat: the account's name, else the guest
 * name, else the seat itself.
 */
export function seatName(s: AdminSeat): string {
  return s.account?.name || s.guest_name || seatNumber(s.seat);
}

/**
 * seatKind is what kind of holder a seat has, in a reader's words: an
 * account, a guest, a Discord seat whose person has not signed in again,
 * a bot and its tier, or an agent and its client.
 */
export function seatKind(s: AdminSeat): string {
  if (s.kind === "bot") return s.bot_tier ? `bot · ${s.bot_tier}` : "bot";
  if (s.kind === "agent") return `agent · ${s.agent_client || "unknown"}`;
  if (s.account) return "account";
  if (s.discord_pending) return "Discord, not signed in yet";
  return "guest";
}

/** seatLabel is a seat in one line: "Ann (account)", "Bot 1 (bot · heuristic)". */
export function seatLabel(s: AdminSeat): string {
  return `${seatName(s)} (${seatKind(s)})`;
}

/**
 * outcomeLine is a table's result: who won, a draw, a close with no
 * result recorded, or where an unfinished table stands.
 */
export function outcomeLine(
  g: Pick<AdminGame, "outcome" | "winner_seat" | "seats" | "state">,
): string {
  switch (g.outcome) {
    case "win": {
      const winner =
        g.winner_seat === undefined ? undefined : g.seats.find((s) => s.seat === g.winner_seat);
      if (winner) return `Won by ${seatName(winner)}`;
      return g.winner_seat === undefined ? "Won" : `Won by ${seatNumber(g.winner_seat)}`;
    }
    case "draw":
      return "Draw";
    case "closed":
      return "Closed, no result";
  }
  if (g.state === "lobby") return "Waiting to start";
  if (g.state === "active") return "In progress";
  return "Ended";
}

/** stateLabel is a table's state in a reader's words. */
export function stateLabel(state: string): string {
  if (state === "lobby") return "waiting";
  if (state === "active") return "running";
  return state;
}

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

/**
 * relativeTime is a past time as "4 min ago" (§7). A time in the future,
 * which only a clock skew between browser and server makes, reads as
 * "just now". Absent reads as "—". Past 60 days it is the date.
 */
export function relativeTime(ms: number | undefined, now: number = Date.now()): string {
  if (ms === undefined || !Number.isFinite(ms) || ms <= 0) return "—";
  const ago = now - ms;
  if (ago < 45_000) return "just now";
  if (ago < HOUR) return `${Math.max(1, Math.round(ago / MINUTE))} min ago`;
  if (ago < DAY) return `${Math.floor(ago / HOUR)} h ago`;
  if (ago < 60 * DAY) return `${Math.floor(ago / DAY)} d ago`;
  return isoDate(ms);
}

function isoDate(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number): string => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** exactTime is the tooltip under a relative time: the local date and time. */
export function exactTime(ms: number | undefined): string {
  if (ms === undefined || !Number.isFinite(ms) || ms <= 0) return "";
  return new Date(ms).toLocaleString();
}

/** isoTime is a time for a <time datetime>. */
export function isoTime(ms: number | undefined): string | undefined {
  if (ms === undefined || !Number.isFinite(ms) || ms <= 0) return undefined;
  return new Date(ms).toISOString();
}

/**
 * avatarSrc is an account's avatar for an <img>. The server serves
 * `/avatars/…` to a session only, and an <img> cannot send a header, so
 * the token rides as ?token=, as api.ts's avatarURL does. Only a
 * same-origin `/avatars/` path is used; anything else is dropped.
 */
export function avatarSrc(url: string | undefined, token: string | undefined): string | null {
  if (!url || !url.startsWith("/avatars/")) return null;
  return token ? `${url}?token=${encodeURIComponent(token)}` : url;
}

/**
 * externalLink is a deck's or an issue's link for an <a href>: an
 * http(s) URL only, so a stored value can never be a `javascript:` link.
 */
export function externalLink(url: string | undefined): string | null {
  if (!url) return null;
  return /^https?:\/\//i.test(url) ? url : null;
}

// --- the accounts list ------------------------------------------------

function desc(a: number | undefined, b: number | undefined): number {
  return (b ?? -1) - (a ?? -1);
}

/**
 * sortAccounts orders the accounts list by `sort`. With none it keeps
 * the server's order (playing now, then the latest play, then the latest
 * sign-in). Ties fall back to the name, so the order is stable.
 */
export function sortAccounts(list: AdminAccount[], sort: AccountSort | undefined): AdminAccount[] {
  if (!sort) return list;
  const byName = (a: AdminAccount, b: AdminAccount): number =>
    a.name.localeCompare(b.name, undefined, { sensitivity: "base" });
  const out = [...list];
  switch (sort) {
    case "name":
      out.sort(byName);
      break;
    case "first_seen":
      out.sort((a, b) => desc(a.first_seen_at, b.first_seen_at) || byName(a, b));
      break;
    case "games":
      out.sort((a, b) => b.games_played - a.games_played || byName(a, b));
      break;
    case "last_played":
      out.sort(
        (a, b) =>
          Number(b.playing_now) - Number(a.playing_now) ||
          desc(a.last_played_at, b.last_played_at) ||
          byName(a, b),
      );
      break;
  }
  return out;
}

/** filterAccounts keeps the accounts whose name contains q, any case. */
export function filterAccounts(list: AdminAccount[], q: string): AdminAccount[] {
  const needle = q.trim().toLocaleLowerCase();
  if (!needle) return list;
  return list.filter((a) => a.name.toLocaleLowerCase().includes(needle));
}

// --- the linked actions (§7) ------------------------------------------

/**
 * archiveConfirm is the question before Archive: the Lobby's copy, and
 * for a running table the line that says archiving ends it.
 */
export function archiveConfirm(g: Pick<AdminGame, "name" | "state">): string {
  const ends =
    g.state === "active"
      ? " This table is running: archiving it ends the game, as /c2-end does, and anyone still connected is dropped."
      : " Bots stop, and anyone still connected is dropped.";
  return `Archive “${g.name}”?${ends} Nothing is deleted: the board, the replay and the invite links are all kept, and you can unarchive it.`;
}

/** revokeConfirm is the question before Revoke sessions, naming the person. */
export function revokeConfirm(name: string): string {
  return `Sign ${name} out everywhere? Every session they hold stops working now, and they stay signed out until they sign in with Discord again. Their account, games and decks are kept.`;
}

/** playmatRemoveConfirm is the question before Remove playmat, naming the person. */
export function playmatRemoveConfirm(name: string): string {
  return `Remove ${name}'s playmat? It disappears from every table they sit at now. They can upload another, and their account, games and decks are kept.`;
}

/** canArchive, canUnarchive and canReplay say which action a table row offers. */
export function canArchive(g: Pick<AdminGame, "archived_at" | "practice">): boolean {
  return !g.archived_at && !g.practice;
}
export function canUnarchive(g: Pick<AdminGame, "archived_at">): boolean {
  return Boolean(g.archived_at);
}
export function canReplay(g: Pick<AdminGame, "state" | "practice">): boolean {
  return g.state === "ended" && !g.practice;
}

// --- Live now's poll schedule (§7) ------------------------------------

/** LIVE_POLL_MS is how often Live now asks while the tab is visible. */
export const LIVE_POLL_MS = 10_000;
/** LIVE_IDLE_MS is how long without input before Live now pauses. */
export const LIVE_IDLE_MS = 10 * MINUTE;

/**
 * PollState is where Live now's refresh stands:
 *   - "polling": visible, and asking every LIVE_POLL_MS;
 *   - "hidden": the tab is hidden, and nothing is asked;
 *   - "paused": no input for LIVE_IDLE_MS; nothing is asked until Resume.
 */
export type PollState = "polling" | "hidden" | "paused";

export interface PollerDeps {
  /** asks GET /admin/live once */
  load: () => unknown;
  now?: () => number;
  /** whether the tab is visible now */
  isVisible?: () => boolean;
  onState?: (s: PollState) => void;
  pollMs?: number;
  idleMs?: number;
}

/**
 * LivePoller is Live now's refresh: one load at start, then one every
 * LIVE_POLL_MS while the tab is visible. A hidden tab stops asking and
 * asks once when it is visible again. After LIVE_IDLE_MS with no pointer
 * or key input (`input()`), it pauses, so a tab left open overnight does
 * not write a log line every 10 seconds; `resume()` is the page's
 * "Resume". The page calls `visibilityChanged()` on `visibilitychange`
 * and `stop()` when it unmounts.
 */
export class LivePoller {
  private timer: ReturnType<typeof setTimeout> | null = null;
  private lastInputAt = 0;
  private current: PollState = "polling";
  private stopped = true;
  private readonly now: () => number;
  private readonly isVisible: () => boolean;
  private readonly pollMs: number;
  private readonly idleMs: number;

  constructor(private readonly deps: PollerDeps) {
    this.now = deps.now ?? Date.now;
    this.isVisible =
      deps.isVisible ??
      (() => typeof document === "undefined" || document.visibilityState !== "hidden");
    this.pollMs = deps.pollMs ?? LIVE_POLL_MS;
    this.idleMs = deps.idleMs ?? LIVE_IDLE_MS;
  }

  get state(): PollState {
    return this.current;
  }

  /** start loads once (if the tab is visible) and begins the schedule. */
  start(): void {
    this.stopped = false;
    this.lastInputAt = this.now();
    if (!this.isVisible()) {
      this.set("hidden");
      return;
    }
    this.run();
  }

  /** stop cancels the schedule; nothing is asked after it. */
  stop(): void {
    this.stopped = true;
    this.clear();
  }

  /** input records pointer or key input on the page. It never un-pauses. */
  input(): void {
    this.lastInputAt = this.now();
  }

  /** resume is "Resume": ask now and poll again. */
  resume(): void {
    if (this.stopped) return;
    this.lastInputAt = this.now();
    if (!this.isVisible()) {
      this.set("hidden");
      return;
    }
    this.run();
  }

  /** visibilityChanged follows the tab: hidden stops, visible asks once. */
  visibilityChanged(): void {
    if (this.stopped) return;
    if (!this.isVisible()) {
      if (this.current === "polling") {
        this.clear();
        this.set("hidden");
      }
      return;
    }
    if (this.current !== "hidden") return;
    // Coming back to the tab is the person looking at it.
    this.lastInputAt = this.now();
    this.run();
  }

  private run(): void {
    this.set("polling");
    this.deps.load();
    this.schedule();
  }

  private tick(): void {
    this.timer = null;
    if (this.stopped || this.current !== "polling") return;
    if (!this.isVisible()) {
      this.set("hidden");
      return;
    }
    if (this.now() - this.lastInputAt >= this.idleMs) {
      this.set("paused");
      return;
    }
    this.deps.load();
    this.schedule();
  }

  private schedule(): void {
    this.clear();
    this.timer = setTimeout(() => this.tick(), this.pollMs);
  }

  private clear(): void {
    if (this.timer !== null) clearTimeout(this.timer);
    this.timer = null;
  }

  private set(s: PollState): void {
    if (s === this.current) return;
    this.current = s;
    this.deps.onState?.(s);
  }
}
