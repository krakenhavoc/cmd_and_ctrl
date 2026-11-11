// draft.ts — a booster draft: the wire types, the HTTP calls, the seat
// token's storage, and the pure rules the draft page runs on (ADR 0222 §5,
// #5438).
//
// Mirrors server/internal/draft (the per-seat view) and
// server/internal/lobby/drafts.go (the routes, all under /draft, where the
// proxies already send the API); docs/lobby.md "Drafts"
// and docs/protocol.md "The draft socket" are the contract. Hand-
// maintained, like draftSets.ts: update both sides in lockstep.
//
// Hidden information (CR 905.1c): a view is built for one seat on the
// server and holds that seat's pack and picks only. Nothing here filters,
// because nothing about another seat's cards is ever sent.

import type { CardImageSize } from "./cardImage";
import { L } from "./labels";
import { costWords } from "./manaSymbol";
import { currentSession, type Session } from "./session";

export const DRAFT_NEW_HASH = "#/draft";

/** The hash of one draft, with the invite when it is a link to share. */
export function draftHash(id: string, invite?: string): string {
  const base = `#/draft/${encodeURIComponent(id)}`;
  return invite ? `${base}?invite=${encodeURIComponent(invite)}` : base;
}

/** The create form's hash with a set chosen already. */
export function draftNewHash(set?: string): string {
  return set ? `${DRAFT_NEW_HASH}?set=${encodeURIComponent(set)}` : DRAFT_NEW_HASH;
}

/** The create form's hash with one of the caller's cubes chosen already. */
export function draftNewCubeHash(cube: string): string {
  return `${DRAFT_NEW_HASH}?cube=${encodeURIComponent(cube)}`;
}

/** The full link a host shares. */
export function draftInviteURL(id: string, invite: string): string {
  return `${location.origin}${location.pathname}${draftHash(id, invite)}`;
}

// ---- The wire ----

export type DraftMode = "pick_one" | "pick_two";
export type DraftState = "lobby" | "picking" | "done";
export type PassDirection = "left" | "right";

export const MIN_SEATS = 2;
export const MAX_SEATS = 8;
export const DEFAULT_SEATS = 8;
/** The longest draft name the server takes, in bytes. */
export const MAX_DRAFT_NAME = 80;
/** The longest guest name the server takes, in bytes. */
export const MAX_GUEST_NAME = 40;

/** One card of a draft. `id` names this copy; two copies have two ids. */
export interface DraftCard {
  id: string;
  /** The printing the image routes take. */
  scryfall_id?: string;
  oracle_id?: string;
  name: string;
  set?: string;
  collector_number?: string;
  rarity?: string;
  /** W, U, B, R, G; empty or absent for a colourless card. */
  colors?: string[];
  /** The printing's colour identity; absent when colourless. A land's colour (#5704). */
  color_identity?: string[];
  /** Brace notation, "{1}{U}"; absent for a land. */
  mana_cost?: string;
  mana_value?: number;
  type_line?: string;
  /** A basic land. */
  basic?: boolean;
}

/** One seat, as everyone may see it: a name and a status, never cards. */
export interface DraftPlayer {
  seat: number;
  name: string;
  host?: boolean;
  bot?: boolean;
  connected?: boolean;
  /** True while this seat has not picked from the pack it holds. */
  picking?: boolean;
  /** How many cards it has drafted. A count, never the cards. */
  picked?: number;
  /** It has a deck for the rounds. A bot always has. */
  built?: boolean;
}

/** Where a draft is once its picks are done (#5440). */
export type DraftPhase = "building" | "rounds" | "finished";

/**
 * A deck built from the viewer's pool: the pool cards it plays, by id,
 * and how many of each basic land it adds. Every pool card it leaves
 * out is the sideboard (CR 100.4b).
 */
export interface DraftDeck {
  main: string[];
  basics?: Record<string, number>;
}

/** The deckbuilding step, while it is open. */
export interface DraftBuild {
  deadline: string;
  now: string;
  /** The whole build timer, in seconds. */
  seconds: number;
  /** The smallest legal deck (CR 100.2b). */
  deck_min: number;
  /** The viewer has submitted a deck of their own. */
  submitted: boolean;
}

/** One match of a round. `a` and `b` are seat numbers. */
export interface DraftMatch {
  a: number;
  /** Absent when `a` sits the round out. */
  b?: number;
  bye?: boolean;
  done: boolean;
  /** The seat that won; absent for a draw. */
  winner?: number;
  wins_a: number;
  wins_b: number;
  /** Two bots did not play this out: the result was a coin toss. */
  coin?: boolean;
}

/** One seat's record across the rounds. */
export interface DraftStanding {
  seat: number;
  wins: number;
  losses: number;
  draws: number;
}

/** The rounds, once they have begun. */
export interface DraftPlay {
  round: number;
  rounds: number;
  /** Every round's matches so far, oldest first. */
  matches: DraftMatch[][];
  standings: DraftStanding[];
  /** The viewer's match in the round in play. */
  yours?: DraftMatch;
  /** The table the viewer's match is played at, once it is open. */
  table?: string;
}

/** The whole draft as one seat may see it: a `draft` frame's payload. */
export interface DraftView {
  id: string;
  name: string;
  /** The set's code; empty for a cube draft. */
  set: string;
  /** The set's name, or the cube's. */
  set_name?: string;
  booster?: string;
  /** A draft of a saved card list (ADR 0222 §8), not of a set's packs. */
  cube?: boolean;
  mode: DraftMode;
  seats: number;
  solo?: boolean;
  state: DraftState;
  seq: number;
  round?: number;
  rounds?: number;
  pick?: number;
  picks_in_round?: number;
  /** How many cards to take from the pack held now; 0 with no pack. */
  take?: number;
  passing?: PassDirection;
  pick_seconds?: number;
  /** When the pick timer runs out, and the server's clock at the frame. */
  deadline?: string;
  now?: string;
  /** The viewer's seat. */
  you: number;
  players: DraftPlayer[];
  /** The pack the viewer holds; absent while waiting for the others. */
  pack?: DraftCard[];
  /** The viewer's own picks, in the order taken. */
  picks?: DraftCard[];
  /** The viewer's whole pool, once the draft is done. */
  pool?: DraftCard[];
  /** The host's invite code, in the lobby of a draft that is not solo. */
  invite?: string;
  /** What follows the picks: absent until `state` is done. */
  phase?: DraftPhase;
  /** The length of the draft's matches: 1 or 3. */
  best_of?: number;
  build?: DraftBuild;
  /** The viewer's own deck, once they have one. */
  deck?: DraftDeck;
  play?: DraftPlay;
}

/** What `GET /draft/{id}?invite=` shows before joining. */
export interface DraftPreview {
  id: string;
  name: string;
  set: string;
  set_name?: string;
  cube?: boolean;
  mode: DraftMode;
  seats: number;
  taken: number;
  state: DraftState;
  host?: string;
}

/** What creating or joining answers with: the seat token and the view. */
export interface DraftSession {
  token: string;
  expires_at: string;
  draft: DraftView;
}

export interface CreateDraftInput {
  /** A draftable set's code. Exactly one of `set` and `cube`. */
  set?: string;
  /** One of the caller's own cubes, by id (ADR 0222 §8); needs a sign-in. */
  cube?: string;
  /** The cards in a cube's packs; the server's default when absent. A set takes none. */
  pack_size?: number;
  mode: DraftMode;
  seats: number;
  name?: string;
  pick_seconds?: number;
  /** The length of the draft's matches: 1 (the default) or 3. */
  best_of?: number;
}

// ---- HTTP ----

export const SERVER_FULL_CODE = "server_full";
export const SET_NOT_DRAFTABLE_CODE = "set_not_draftable";

/** A refused draft request: the server's sentence, its status and its code. */
export class DraftApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public code?: string,
  ) {
    super(message);
    this.name = "DraftApiError";
  }
}

/**
 * isDraftGone reports whether err says the draft, or this seat in it, is
 * no longer there: 404 (unknown draft, wrong invite), 401 (no valid seat
 * token) or 403 (a token of another draft).
 */
export function isDraftGone(err: unknown): boolean {
  return err instanceof DraftApiError && [401, 403, 404].includes(err.status);
}

/** The sentence to show for a failed draft request. */
export function draftErrorMessage(err: unknown, fallback: string): string {
  return err instanceof DraftApiError && err.message ? err.message : fallback;
}

// Plain `fetch`, not `authFetch`: these routes are public, and authFetch
// reads any 401 as an expired sign-in and clears it. A 401 here is about
// the seat token, which is not the session. On create and join the server
// recognises a signed-in person by their cookie (`credentials`) or, when
// the cookie has lapsed but the session in memory is still good, by that
// session as a Bearer (signedInBearer), as every authFetch call sends it.
// A session the server no longer accepts is ignored there: the caller is
// then a visitor.
async function draftFetch(
  path: string,
  init: { method?: string; body?: unknown; token?: string; signal?: AbortSignal } = {},
): Promise<Response> {
  const headers: Record<string, string> = { Accept: "application/json" };
  if (init.body !== undefined) headers["Content-Type"] = "application/json";
  if (init.token) headers["Authorization"] = `Bearer ${init.token}`;
  const res = await fetch(path, {
    method: init.method ?? "GET",
    headers,
    credentials: "same-origin",
    body: init.body === undefined ? undefined : JSON.stringify(init.body),
    signal: init.signal,
  });
  if (res.ok) return res;
  let message = `${res.status} ${res.statusText}`.trim();
  let code: string | undefined;
  try {
    const body = (await res.json()) as { error?: unknown; code?: unknown };
    if (typeof body.error === "string" && body.error) message = body.error;
    if (typeof body.code === "string") code = body.code;
  } catch {
    // not JSON — keep the status line
  }
  throw new DraftApiError(res.status, message, code);
}

/** The caller's own session token, if they have one: who is asking. */
function signedInBearer(): string | undefined {
  return currentSession()?.token || undefined;
}

const draftPath = (id: string, rest = ""): string => `/draft/${encodeURIComponent(id)}${rest}`;

/** `POST /draft`. The caller keeps the token (rememberDraftSeat). */
export async function createDraft(input: CreateDraftInput): Promise<DraftSession> {
  const body: Record<string, unknown> = { mode: input.mode, seats: input.seats };
  if (input.cube) {
    body.cube = input.cube;
    if (input.pack_size !== undefined) body.pack_size = input.pack_size;
  } else {
    body.set = input.set ?? "";
  }
  const name = input.name?.trim();
  if (name) body.name = name;
  if (input.pick_seconds !== undefined) body.pick_seconds = input.pick_seconds;
  if (input.best_of === 3) body.best_of = 3;
  const res = await draftFetch("/draft", { method: "POST", body, token: signedInBearer() });
  return (await res.json()) as DraftSession;
}

/** `GET /draft/{id}`, by the invite or by a seat token of that draft. */
export async function previewDraft(
  id: string,
  by: { invite?: string; token?: string },
  signal?: AbortSignal,
): Promise<DraftPreview> {
  const query = by.invite ? `?invite=${encodeURIComponent(by.invite)}` : "";
  const res = await draftFetch(draftPath(id, query), { token: by.token, signal });
  return (await res.json()) as DraftPreview;
}

/** `POST /draft/{id}/join`. A signed-in person's `name` is ignored. */
export async function joinDraft(id: string, invite: string, name?: string): Promise<DraftSession> {
  const body: Record<string, unknown> = { invite };
  if (name?.trim()) body.name = name.trim();
  const res = await draftFetch(draftPath(id, "/join"), {
    method: "POST",
    body,
    token: signedInBearer(),
  });
  return (await res.json()) as DraftSession;
}

/** `POST /draft/{id}/start`, the host's seat token. */
export async function startDraft(id: string, token: string): Promise<DraftView> {
  const res = await draftFetch(draftPath(id, "/start"), { method: "POST", token });
  return ((await res.json()) as { draft: DraftView }).draft;
}

/** `POST /draft/{id}/leave`, a seat token. */
export async function leaveDraft(id: string, token: string): Promise<void> {
  await draftFetch(draftPath(id, "/leave"), { method: "POST", token });
}

/**
 * `POST /draft/{id}/match`, a seat token: sits the caller at the table
 * of their match in the round in play. The answer is a table seat's
 * session, as a join's is, and the server sets it as the cookie the
 * game page reads. The draft seat token is not spent.
 */
export async function seatDraftMatch(id: string, token: string): Promise<Session> {
  const res = await draftFetch(draftPath(id, "/match"), { method: "POST", token });
  const body = (await res.json()) as {
    token: string;
    expires_at: string;
    principal: Session["principal"];
    player_id?: string;
    game?: { id?: string };
  };
  return {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
    playerID: body.player_id,
    gameID: body.game?.id,
  };
}

/** The drag type a pool card carries on the deckbuilding screen: its draft id. */
export const DRAG_TYPE = "application/x-c2-draft-card";

// The way back from a match's table to its draft. The game page knows a
// table by its id alone, so the draft page notes which draft a table
// belongs to as it sends the player there. Session storage: it is this
// tab's errand, and it is gone with the tab.
const MATCH_KEY = "c2.draftMatch.";

/** Notes that table `gameID` is a match of draft `draftID`. */
export function rememberDraftMatch(gameID: string, draftID: string): void {
  try {
    sessionStorage.setItem(MATCH_KEY + gameID, draftID);
  } catch {
    // storage is off: the game page falls back to the lobby
  }
}

/** The draft a table is a match of, or null. */
export function draftOfMatch(gameID: string): string | null {
  try {
    return sessionStorage.getItem(MATCH_KEY + gameID);
  } catch {
    return null;
  }
}

/** The draft socket's URL for a seat: `/draft/{id}/ws?token=`. */
export function draftSocketURL(id: string, token: string, loc: Location = location): string {
  const scheme = loc.protocol === "https:" ? "wss://" : "ws://";
  return `${scheme}${loc.host}/draft/${encodeURIComponent(id)}/ws?token=${encodeURIComponent(token)}`;
}

/**
 * A draft card's image, or null for a card with no printing to draw
 * (it is then drawn as its name).
 *
 * NOT scryfallImageURL from cardImage.ts: /cards/<id>/image wants a
 * session cookie, and a guest drafter has none (the seat token sits in
 * localStorage, and an <img> cannot send it). Nor the catalogue's image
 * route, which serves each card's representative printing only. The
 * public /draft/image/<id> serves exactly the printings a draftable
 * set's packs can contain, basic lands included.
 *
 * Face 0 omits the `face` parameter, as cardImage.ts and catalog.ts do:
 * the caches key on the whole URL.
 */
export function draftImageURL(
  card: Pick<DraftCard, "scryfall_id">,
  size: CardImageSize = "normal",
  face = 0,
): string | null {
  if (!card.scryfall_id) return null;
  const base = `/draft/image/${card.scryfall_id}?size=${size}`;
  return face > 0 ? `${base}&face=${face}` : base;
}

// ---- The seat token ----
//
// The token is not a cookie: it sits in localStorage under the draft's
// id, so a reload or a return visit takes the same seat and the same
// picks. Every access is wrapped, because storage can be off or full.

const SEAT_KEY_PREFIX = "cmdctrl.draft.";

export interface StoredSeat {
  token: string;
  /** ISO 8601; absent when the server named none. */
  expiresAt?: string;
}

const seatKey = (id: string): string => SEAT_KEY_PREFIX + id;

function parseSeat(raw: string | null, now: number): StoredSeat | null {
  if (!raw) return null;
  try {
    const v = JSON.parse(raw) as Partial<StoredSeat> | null;
    if (!v || typeof v.token !== "string" || !v.token) return null;
    if (typeof v.expiresAt === "string") {
      const t = Date.parse(v.expiresAt);
      if (Number.isFinite(t) && t <= now) return null;
      return { token: v.token, expiresAt: v.expiresAt };
    }
    return { token: v.token };
  } catch {
    return null;
  }
}

/** Keep the seat a create or a join answered with. */
export function rememberDraftSeat(id: string, s: Pick<DraftSession, "token" | "expires_at">): void {
  try {
    const seat: StoredSeat = { token: s.token, expiresAt: s.expires_at };
    localStorage.setItem(seatKey(id), JSON.stringify(seat));
  } catch {
    // storage is off: the seat lasts as long as this page does
  }
}

/** The seat this browser holds in a draft, or null. An expired one is forgotten. */
export function storedDraftSeat(id: string, now: number = Date.now()): StoredSeat | null {
  try {
    const raw = localStorage.getItem(seatKey(id));
    const seat = parseSeat(raw, now);
    if (raw && !seat) localStorage.removeItem(seatKey(id));
    return seat;
  } catch {
    return null;
  }
}

/** Forget a draft's seat: it left, or the server says it is gone. */
export function forgetDraftSeat(id: string): void {
  try {
    localStorage.removeItem(seatKey(id));
  } catch {
    // nothing to forget
  }
}

/** The ids of the drafts this browser holds a live seat in. */
export function storedDraftIDs(now: number = Date.now()): string[] {
  const ids: string[] = [];
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (!key?.startsWith(SEAT_KEY_PREFIX)) continue;
      if (parseSeat(localStorage.getItem(key), now)) ids.push(key.slice(SEAT_KEY_PREFIX.length));
    }
  } catch {
    return [];
  }
  return ids;
}

// ---- The create form ----

/**
 * The pick mode offered first for a number of seats: pick two at four,
 * pick one otherwise. Either is allowed at any size.
 */
export function defaultPickMode(seats: number): DraftMode {
  return seats === 4 ? "pick_two" : "pick_one";
}

/** The seat counts the form offers. */
export function seatChoices(): number[] {
  return Array.from({ length: MAX_SEATS - MIN_SEATS + 1 }, (_, i) => MIN_SEATS + i);
}

/** A pick mode's name. */
export function pickModeName(mode: DraftMode | string | undefined): string {
  return mode === "pick_two" ? L.draftPickTwo : L.draftPickOne;
}

// ---- The pick ----

/**
 * The selection after a card is clicked. Clicking a chosen card lets it
 * go. With room left the card is added. With `take` cards chosen already
 * the oldest choice gives way, so with one card to take a click simply
 * moves the choice, and with two the pair is always the last two clicked.
 */
export function toggleSelection(selected: readonly string[], id: string, take: number): string[] {
  if (take <= 0) return [];
  if (selected.includes(id)) return selected.filter((s) => s !== id);
  const next = [...selected, id];
  return next.length > take ? next.slice(next.length - take) : next;
}

/** The selection with every card the pack no longer holds dropped, and no more than `take`. */
export function pruneSelection(
  selected: readonly string[],
  pack: readonly DraftCard[] | undefined,
  take: number,
): string[] {
  const held = new Set((pack ?? []).map((c) => c.id));
  const kept = selected.filter((id) => held.has(id));
  return take > 0 ? kept.slice(Math.max(0, kept.length - take)) : [];
}

/** True when the selection is exactly what the server will take. */
export function canConfirm(
  selected: readonly string[],
  pack: readonly DraftCard[] | undefined,
  take: number,
): boolean {
  if (take <= 0 || selected.length !== take) return false;
  if (new Set(selected).size !== selected.length) return false;
  const held = new Set((pack ?? []).map((c) => c.id));
  return selected.every((id) => held.has(id));
}

/** How many cards the viewer must take now; 0 when they hold no pack. */
export function takeOf(view: DraftView): number {
  if (view.state !== "picking" || !view.pack || view.pack.length === 0) return 0;
  const take = view.take ?? 0;
  return Math.max(0, Math.min(take, view.pack.length));
}

// ---- The seats ----

/** The viewer's own seat row, or undefined before it is listed. */
export function self(view: DraftView): DraftPlayer | undefined {
  return view.players.find((p) => p.seat === view.you);
}

/** True when the viewer hosts the draft. */
export function isHost(view: DraftView): boolean {
  return self(view)?.host === true;
}

/** Seats nobody has taken yet. Only the lobby has any. */
export function openSeats(view: DraftView): number {
  return view.state === "lobby" ? Math.max(0, view.seats - view.players.length) : 0;
}

/** The seats that have not picked from the pack they hold, in seat order. */
export function stillPicking(view: DraftView): DraftPlayer[] {
  if (view.state !== "picking") return [];
  return view.players.filter((p) => p.picking === true).sort((a, b) => a.seat - b.seat);
}

/** The other seats the viewer is waiting on. */
export function waitingOn(view: DraftView): DraftPlayer[] {
  return stillPicking(view).filter((p) => p.seat !== view.you);
}

/** A person whose socket is closed: their picks are being made for them. */
export function isAway(p: DraftPlayer): boolean {
  return !p.bot && p.connected === false;
}

export type SeatStatus = "picking" | "picked" | "away" | "seated";

/** What a seat is doing, for the seat list. */
export function seatStatus(view: DraftView, p: DraftPlayer): SeatStatus {
  if (isAway(p)) return "away";
  if (view.state !== "picking") return "seated";
  return p.picking ? "picking" : "picked";
}

/** The word shown for a seat's status. */
export function seatStatusWord(status: SeatStatus): string {
  switch (status) {
    case "picking":
      return "picking";
    case "picked":
      return "picked";
    case "away":
      return "away";
    case "seated":
      return "";
  }
}

const nameList = (names: readonly string[]): string =>
  names.length <= 1
    ? (names[0] ?? "")
    : `${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}`;

/** "Waiting for Ian and Bot 3." Up to three names, then a count. */
export function waitingLine(view: DraftView): string {
  const others = waitingOn(view);
  if (others.length === 0) return "Everyone has picked. The packs are being passed…";
  if (others.length > 3) return `Waiting for ${others.length} players.`;
  return `Waiting for ${nameList(others.map((p) => p.name))}.`;
}

/**
 * The seat the viewer's pack goes to next (CR 905.1b): the player to the
 * left is the next higher seat, wrapping round. Undefined outside a pick.
 */
export function passTarget(view: DraftView): DraftPlayer | undefined {
  if (view.state !== "picking" || !view.passing || view.seats < 2) return undefined;
  const step = view.passing === "left" ? 1 : -1;
  const seat = (((view.you + step) % view.seats) + view.seats) % view.seats;
  return view.players.find((p) => p.seat === seat);
}

/** "Pack 2 of 3 · pick 5 of 14 · passing right". "" outside a pick. */
export function progressLine(view: DraftView): string {
  if (view.state !== "picking" || !view.round) return "";
  const parts = [`Pack ${view.round} of ${view.rounds ?? view.round}`];
  if (view.pick) parts.push(`pick ${view.pick} of ${view.picks_in_round ?? view.pick}`);
  if (view.passing) parts.push(`passing ${view.passing}`);
  return parts.join(" · ");
}

// ---- The pick timer ----

/** The last seconds of a pick, when the timer turns urgent. */
export const URGENT_SECONDS = 10;

/**
 * Milliseconds left on the pick timer when the frame was built, by the
 * server's clock: `deadline` less `now`, never the viewer's own clock.
 * Null with no timer running (the lobby, a finished draft, a bad date).
 */
export function msLeftAtFrame(view: Pick<DraftView, "deadline" | "now">): number | null {
  if (!view.deadline || !view.now) return null;
  const deadline = Date.parse(view.deadline);
  const now = Date.parse(view.now);
  if (!Number.isFinite(deadline) || !Number.isFinite(now)) return null;
  return Math.max(0, deadline - now);
}

/**
 * Whole seconds left, `elapsedMs` after the frame arrived. The count
 * runs on time passed since the frame, so a viewer whose clock is wrong
 * still sees the server's timer.
 */
export function secondsLeft(
  view: Pick<DraftView, "deadline" | "now">,
  elapsedMs: number,
): number | null {
  const at = msLeftAtFrame(view);
  if (at === null) return null;
  return Math.max(0, Math.ceil((at - Math.max(0, elapsedMs)) / 1000));
}

/** "1:05", "0:09". */
export function clockText(seconds: number): string {
  const s = Math.max(0, Math.floor(seconds));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
}

/** True in the last seconds of a pick. */
export function isUrgent(seconds: number | null): boolean {
  return seconds !== null && seconds <= URGENT_SECONDS;
}

// The moments a screen reader hears the timer: not every second.
const TIMER_MARKS = [30, URGENT_SECONDS];

/**
 * What to announce as the timer moves from `prev` to `next` seconds:
 * one line as it crosses 30 and 10, nothing otherwise.
 */
export function timerAnnouncement(prev: number | null, next: number | null): string {
  if (prev === null || next === null || next >= prev) return "";
  for (const mark of TIMER_MARKS) {
    if (prev > mark && next <= mark && next > 0) return `${mark} seconds left to pick.`;
  }
  return "";
}

/** "Pick 1 card" / "Pick 2 cards". */
export function pickCount(take: number): string {
  return L.draftConfirmPick(take);
}

// ---- The picks, grouped ----

export type PoolGroupKey = "W" | "U" | "B" | "R" | "G" | "multi" | "colorless" | "land";

/** The order the groups are shown in: WUBRG, many colours, none, lands. */
export const POOL_GROUP_ORDER: readonly PoolGroupKey[] = [
  "W",
  "U",
  "B",
  "R",
  "G",
  "multi",
  "colorless",
  "land",
];

const COLOURS = new Set(["W", "U", "B", "R", "G"]);

/** True for a land: its front face's type line says so. */
export function isLand(card: Pick<DraftCard, "type_line" | "basic">): boolean {
  if (card.basic) return true;
  const front = (card.type_line ?? "").split("//")[0];
  return /\bLand\b/.test(front);
}

/** The group a card sits in. A land is a land whatever its colours. */
export function poolGroupOf(card: DraftCard): PoolGroupKey {
  if (isLand(card)) return "land";
  const colours = [...new Set((card.colors ?? []).filter((c) => COLOURS.has(c)))];
  if (colours.length === 0) return "colorless";
  if (colours.length > 1) return "multi";
  return colours[0] as PoolGroupKey;
}

/** A group's heading. */
export function poolGroupName(key: PoolGroupKey): string {
  switch (key) {
    case "W":
      return L.draftGroupWhite;
    case "U":
      return L.draftGroupBlue;
    case "B":
      return L.draftGroupBlack;
    case "R":
      return L.draftGroupRed;
    case "G":
      return L.draftGroupGreen;
    case "multi":
      return L.draftGroupMulticolour;
    case "colorless":
      return L.draftGroupColourless;
    case "land":
      return L.draftGroupLands;
  }
}

/** One mana value's cards within a group, by name. */
export interface PoolValueRow {
  manaValue: number;
  cards: DraftCard[];
}

export interface PoolGroup {
  key: PoolGroupKey;
  name: string;
  /** How many cards the group holds. */
  count: number;
  /** Every card, by mana value and then by name. */
  cards: DraftCard[];
  /** The same cards in rows of one mana value each, lowest first. */
  byValue: PoolValueRow[];
}

const valueOf = (c: DraftCard): number =>
  typeof c.mana_value === "number" && Number.isFinite(c.mana_value) ? c.mana_value : 0;

/**
 * The cards grouped by colour, and by mana value inside each group. A
 * group with no card is left out; the order is POOL_GROUP_ORDER.
 */
export function groupPool(cards: readonly DraftCard[] | undefined): PoolGroup[] {
  const buckets = new Map<PoolGroupKey, DraftCard[]>();
  for (const card of cards ?? []) {
    const key = poolGroupOf(card);
    const list = buckets.get(key);
    if (list) list.push(card);
    else buckets.set(key, [card]);
  }
  const groups: PoolGroup[] = [];
  for (const key of POOL_GROUP_ORDER) {
    const list = buckets.get(key);
    if (!list) continue;
    const sorted = [...list].sort(
      (a, b) => valueOf(a) - valueOf(b) || a.name.localeCompare(b.name) || a.id.localeCompare(b.id),
    );
    const byValue: PoolValueRow[] = [];
    for (const card of sorted) {
      const last = byValue[byValue.length - 1];
      if (last && last.manaValue === valueOf(card)) last.cards.push(card);
      else byValue.push({ manaValue: valueOf(card), cards: [card] });
    }
    groups.push({ key, name: poolGroupName(key), count: sorted.length, cards: sorted, byValue });
  }
  return groups;
}

/** A card and how many copies of it sit together in a list. */
export interface StackedCard {
  card: DraftCard;
  count: number;
}

/** Neighbouring copies of one card folded into a row with a count. */
export function stackCopies(cards: readonly DraftCard[]): StackedCard[] {
  const out: StackedCard[] = [];
  for (const card of cards) {
    const last = out[out.length - 1];
    if (last && last.card.name === card.name) last.count += 1;
    else out.push({ card, count: 1 });
  }
  return out;
}

/** "1 card" / "42 cards". */
export function cardCount(n: number): string {
  return `${n} ${n === 1 ? "card" : "cards"}`;
}

// ---- Words for a screen reader ----

/**
 * A card's accessible name: its name first, then what a sighted player
 * reads off the image at a glance (cost, type, rarity).
 */
export function cardAccessibleName(card: DraftCard): string {
  const parts = [card.name];
  const cost = card.mana_cost ? costWords(card.mana_cost) : "";
  if (cost) parts.push(`costs ${cost}`);
  if (card.type_line) parts.push(card.type_line);
  if (card.rarity) parts.push(card.rarity);
  return parts.join(", ");
}

const packKey = (v: DraftView): string =>
  v.pack && v.pack.length > 0 ? `${v.round ?? 0}:${v.pick ?? 0}` : "";

/**
 * What to announce, once, when one view replaces another: the start, a
 * new pack, a pick that landed, the end. "" when nothing a listener must
 * know has changed (a seat connecting, someone else picking).
 */
export function viewAnnouncement(prev: DraftView | null, next: DraftView): string {
  if (next.state === "done") {
    return prev?.state === "done" ? "" : `${L.draftFinished}.`;
  }
  if (next.state !== "picking") return "";
  const before = prev && prev.state === "picking" ? packKey(prev) : "";
  const after = packKey(next);
  if (after && after !== before) {
    const started = prev?.state === "lobby" ? "The draft has started. " : "";
    return `${started}${progressLine(next)}. ${pickCount(takeOf(next))}.`;
  }
  if (!after && before) return `Pick made. ${waitingLine(next)}`;
  return "";
}
