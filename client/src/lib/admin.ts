// Who the client treats as an admin (ADR 0110 §3 item 4, ADR 0112 §2).
//
// Two kinds of session are admins on the server:
//   - the shared admin token's session (`role: "admin"`, from
//     POST /admin/login), and
//   - a signed-in person whose Discord ID is on the server's allowlist
//     (CMDCTRL_DISCORD_ADMIN_USER_IDS) AND who has admin mode on (ADR
//     0112 §2). Their session is an ordinary identified, player or
//     spectator session; nothing in it says "admin". The server
//     computes that per request, and GET /me reports it.
//
// Being on the list only makes a person ABLE to be an admin. Admin mode
// starts off ("player mode"), the Admin chip in the header turns it on
// for 12 hours (owner answer 1), and it lapses by itself after that.
//
// So the client asks /me per installed session that carries a user,
// and keeps the answer on the session: `admin_allowed`, `admin_mode`,
// `admin_mode_ends_at` (and `admin`, the server's effective answer).
// Every admin affordance reads isAdmin(session), never
// `role === "admin"`, so a seated admin in admin mode sees the admin UI
// exactly as the token does, and the same person in player mode sees
// what every other player sees.
//
// This is a mirror of the server, never the gate: every admin route and
// every admin socket binding is decided server-side, and a stale `true`
// here only shows a control whose click is refused.
//
// /me is asked again (ADR 0112 §2 item 9):
//   - once per installed session per page load (App.svelte);
//   - after a switch: PUT /me/admin-mode answers with the same fields,
//     and the session takes them from that answer;
//   - on a 4001 "admin mode changed" close (ws.ts, afterAdminModeChanged);
//   - when a hidden tab becomes visible, for an allowlisted person;
//   - when admin_mode_ends_at passes (armAdminLapse).

import { type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { signedInUserID } from "./myGames";
import { authFetch, currentSession, session, type Session } from "./session";

/** AdminModeFields is what GET /me and PUT /me/admin-mode say about admin. */
export interface AdminModeFields {
  admin?: boolean;
  admin_allowed?: boolean;
  admin_mode?: boolean;
  admin_mode_ends_at?: number;
}

// hasUser reports whether s carries a person. A session with no user
// (or the nil one) cannot be on the allowlist: the server requires a
// user (a reclaim ticket carries a seat's Discord ID but no user, and is
// never admin).
function hasUser(s: Session): boolean {
  return signedInUserID(s) !== null;
}

/**
 * adminModeOn reports whether s, a person's session, has admin mode on
 * at `now`: on the allowlist, switched on, and not yet lapsed. A mode
 * with no end time is treated as lapsed: the server always sends one
 * while the mode is on, so a missing one is a stale or hand-made
 * session, and the safe reading of it is "off".
 */
export function adminModeOn(s: Session | null | undefined, now: number = Date.now()): boolean {
  if (!s || !hasUser(s)) return false;
  if (s.admin_allowed !== true || s.admin_mode !== true) return false;
  const ends = s.admin_mode_ends_at;
  return typeof ends === "number" && ends > now;
}

/**
 * isAdmin reports whether s is an admin session right now: the shared
 * token, or an allowlisted person in admin mode that has not lapsed.
 */
export function isAdmin(s: Session | null | undefined, now: number = Date.now()): boolean {
  if (!s) return false;
  if (s.principal.role === "admin") return true;
  return adminModeOn(s, now);
}

/**
 * isAdminAllowed reports whether s is an allowlisted person: someone
 * who may switch admin mode on, in either mode. The token is not (it
 * has no mode).
 */
export function isAdminAllowed(s: Session | null | undefined): boolean {
  return !!s && s.principal.role !== "admin" && hasUser(s) && s.admin_allowed === true;
}

// --- the chip ---------------------------------------------------------

/**
 * AdminChip is what the header's account menu shows in the chip's slot
 * (ADR 0112 §2 item 9): the switch for an allowlisted person, the
 * static "Admin token" badge for the token, or nothing.
 */
export type AdminChip =
  | { kind: "token" }
  | { kind: "switch"; on: boolean; endsAt: number | null }
  | null;

export function adminChipFor(s: Session | null | undefined, now: number = Date.now()): AdminChip {
  if (!s) return null;
  if (s.principal.role === "admin") return { kind: "token" };
  if (!isAdminAllowed(s)) return null;
  const on = adminModeOn(s, now);
  return { kind: "switch", on, endsAt: on ? (s.admin_mode_ends_at ?? null) : null };
}

/** clockTime is a Unix-millisecond time as the local "HH:MM". */
export function clockTime(ms: number): string {
  const d = new Date(ms);
  if (Number.isNaN(d.getTime())) return "";
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

/** adminChipLabel is the chip's text: "Admin · until 23:40", or "Player". */
export function adminChipLabel(on: boolean, endsAt: number | null): string {
  if (!on) return "Player";
  return endsAt === null ? "Admin" : `Admin · until ${clockTime(endsAt)}`;
}

/** adminChipTitle is the chip's tooltip: it names the other mode. */
export function adminChipTitle(on: boolean): string {
  return on
    ? "Admin mode is on. Switch to player mode"
    : "Player mode is on. Switch to admin mode for 12 hours";
}

/** adminSwitchLabel is the in-game ⋯ menu's item for the same switch. */
export function adminSwitchLabel(on: boolean): string {
  return on ? "Switch to player mode" : "Switch to admin mode";
}

// --- asking /me -------------------------------------------------------

// The tokens already asked about during this page load. A session is
// asked once by App.svelte; a reload asks again, so an ID taken off the
// list stops showing admin UI on the next visit. The other reasons to
// ask (a switch, a 4001, a visible tab, a lapse) ask regardless.
const asked = new Set<string>();

/**
 * needsAdminCheck reports whether s should be asked about: a session
 * with a user that has not been asked during this page load. The token
 * session is already known to be admin, and a session with no user
 * never is.
 */
export function needsAdminCheck(s: Session | null | undefined): s is Session {
  if (!s || s.principal.role === "admin" || !hasUser(s)) return false;
  return !asked.has(s.token);
}

// pickAdminFields keeps the admin answer from a /me or PUT body, typed.
// Missing or mistyped fields read as "no": the safe answer.
function pickAdminFields(body: unknown): Required<Omit<AdminModeFields, "admin_mode_ends_at">> & {
  admin_mode_ends_at: number | undefined;
} {
  const b = (body ?? {}) as Record<string, unknown>;
  const ends = b.admin_mode_ends_at;
  return {
    admin: b.admin === true,
    admin_allowed: b.admin_allowed === true,
    admin_mode: b.admin_mode === true,
    admin_mode_ends_at: typeof ends === "number" && Number.isFinite(ends) ? ends : undefined,
  };
}

// recordAdminFields puts `fields` on the installed session, if it is
// still the one with `token`. Only the fields given are written.
function recordAdminFields(token: string, fields: AdminModeFields): void {
  session.update((cur) => {
    if (!cur || cur.token !== token) return cur;
    const next: Session = { ...cur, ...fields };
    if (!("admin_mode_ends_at" in fields) || fields.admin_mode_ends_at === undefined) {
      delete next.admin_mode_ends_at;
    }
    if (
      cur.admin === next.admin &&
      cur.admin_allowed === next.admin_allowed &&
      cur.admin_mode === next.admin_mode &&
      cur.admin_mode_ends_at === next.admin_mode_ends_at
    ) {
      return cur;
    }
    return next;
  });
}

/**
 * loadAdminStatus asks GET /me about s and records the answer on the
 * session, if s is still the installed one. It resolves true when /me
 * answered. A failure leaves the session as it was and resolves false:
 * a network blip must not take the admin UI away, and an expired
 * session is authFetch's business (it clears it on a 401).
 */
export async function loadAdminStatus(s: Session): Promise<boolean> {
  asked.add(s.token);
  let fields: ReturnType<typeof pickAdminFields>;
  try {
    const res = await authFetch("/me");
    fields = pickAdminFields(await res.json());
  } catch {
    return false;
  }
  recordAdminFields(s.token, fields);
  return true;
}

/**
 * refreshAdminStatus asks /me again about the installed session, if it
 * has a user, whether or not it was asked before. Resolves true when
 * /me answered (or there is nothing to ask), false when it failed.
 */
export async function refreshAdminStatus(): Promise<boolean> {
  const s = currentSession();
  if (!s || s.principal.role === "admin" || !hasUser(s)) return true;
  return loadAdminStatus(s);
}

/**
 * switchAdminMode is the chip: PUT /me/admin-mode {on}, and the answer
 * goes onto the session. A refusal throws (authFetch's LobbyApiError,
 * carrying the server's message) and leaves the session as it was.
 */
export async function switchAdminMode(on: boolean): Promise<void> {
  const s = currentSession();
  const res = await authFetch("/me/admin-mode", {
    method: "PUT",
    body: JSON.stringify({ on }),
  });
  const fields = pickAdminFields(await res.json());
  if (!s) return;
  // Only an allowlisted person gets a 200, so the answer implies it.
  recordAdminFields(s.token, { ...fields, admin_allowed: true });
}

// --- the lapse (owner answer 1) ---------------------------------------

// LAPSE_RECHECK_MIN_MS is the shortest gap between two re-asks caused
// by a lapse. The server's sweeper runs once a minute; a client whose
// clock runs ahead of the server's would otherwise see the end time
// pass, ask, hear "still on, ends at <a time already past here>", and
// ask again at once, forever.
export const LAPSE_RECHECK_MIN_MS = 60_000;

let lapseTimer: ReturnType<typeof setTimeout> | null = null;
let lapseArmedFor: { token: string; endsAt: number } | null = null;
let lastLapseAsk = 0;

/**
 * armAdminLapse (re)arms the timer that ends admin mode on the client
 * when s's admin_mode_ends_at passes: the session drops admin mode
 * there and then (the server's answer from that moment), and /me is
 * asked again to confirm. Call it with every installed session; a
 * session with no mode on disarms it.
 */
export function armAdminLapse(s: Session | null | undefined): void {
  const ends = s && s.admin_mode === true ? s.admin_mode_ends_at : undefined;
  if (!s || typeof ends !== "number") {
    disarmAdminLapse();
    return;
  }
  if (lapseArmedFor && lapseArmedFor.token === s.token && lapseArmedFor.endsAt === ends) return;
  disarmAdminLapse();
  const now = Date.now();
  const earliest = lastLapseAsk + LAPSE_RECHECK_MIN_MS;
  const delay = Math.max(0, ends - now, ends <= now ? earliest - now : 0);
  lapseArmedFor = { token: s.token, endsAt: ends };
  const token = s.token;
  lapseTimer = setTimeout(() => {
    lapseTimer = null;
    lapseArmedFor = null;
    const cur = currentSession();
    if (!cur || cur.token !== token || cur.admin_mode_ends_at !== ends) return;
    lastLapseAsk = Date.now();
    recordAdminFields(token, { admin: false, admin_mode: false, admin_mode_ends_at: undefined });
    void loadAdminStatus(cur);
  }, delay);
}

function disarmAdminLapse(): void {
  if (lapseTimer !== null) clearTimeout(lapseTimer);
  lapseTimer = null;
  lapseArmedFor = null;
}

/**
 * onVisibleAgain asks /me again when a hidden tab becomes visible, for
 * an allowlisted person: another tab or device may have switched the
 * mode meanwhile.
 */
export function onVisibleAgain(): void {
  if (typeof document !== "undefined" && document.visibilityState !== "visible") return;
  if (!isAdminAllowed(currentSession())) return;
  void refreshAdminStatus();
}

// --- a 4001 at the table (ADR 0112 §2 item 5) -------------------------

/** What GameClient does after a 4001 "admin mode changed" close. */
export type AdminModeVerdict = "reconnect" | "leave" | "retry";

/**
 * PLAYER_MODE_NOT_YOURS is what the Lobby says to a person sent there
 * from a table whose connection only an admin may hold.
 */
export const PLAYER_MODE_NOT_YOURS = "Player mode is on; this table isn't yours.";

/** adminNotice is a one-shot line for the Lobby, cleared once shown. */
export const adminNotice: Writable<string> = guardedWritable("", "adminNotice");

/**
 * ownBindingFor reports whether s may connect to gameID with no admin
 * rights at all: a player session for this game (its own seat) or a
 * spectator session for this game. Every other binding (another game,
 * another seat, the seatless unfiltered view, an `identified` session)
 * is one only an admin may hold, and the server refuses its upgrade.
 */
export function ownBindingFor(s: Session | null | undefined, gameID: string): boolean {
  if (!s) return false;
  const game = s.gameID ?? s.principal.game_id;
  if (game !== gameID) return false;
  if (s.principal.role === "player") return Boolean(s.playerID ?? s.principal.player_id);
  return s.principal.role === "spectator";
}

/**
 * adminModeVerdict is the decision after /me has answered: reconnect
 * when the session may hold a binding for this table (as an admin, or
 * its own seat or spectator session), leave for the Lobby otherwise.
 */
export function adminModeVerdict(
  s: Session | null | undefined,
  gameID: string,
  now: number = Date.now(),
): Exclude<AdminModeVerdict, "retry"> {
  if (isAdmin(s, now)) return "reconnect";
  return ownBindingFor(s, gameID) ? "reconnect" : "leave";
}

/**
 * afterAdminModeChanged is GameClient's 4001 handler for gameID: ask
 * /me first, then decide. A /me that did not answer is "retry": the
 * client waits a rung of its backoff ladder and asks again, and never
 * dials a binding it cannot know it may hold.
 */
export async function afterAdminModeChanged(gameID: string): Promise<AdminModeVerdict> {
  if (!(await refreshAdminStatus())) return currentSession() ? "retry" : "leave";
  return adminModeVerdict(currentSession(), gameID);
}

/** Test seam: forget which sessions were asked, and disarm the lapse. */
export function resetAdminChecksForTest(): void {
  asked.clear();
  disarmAdminLapse();
  lastLapseAsk = 0;
}
