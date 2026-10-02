// practiceTable.ts — the tutorial's practice table, client half (ADR
// 0076 §2.2, #1078): opening it, and putting everything back when the
// player leaves it by any door.
//
// Opening it swaps two things the player owns for the tutorial's
// duration:
//
//   - Four settings are FORCED: strictMana off, autoPassPriority off
//     (so priority visibly reaches the player), tableLayout quadrant,
//     cardSize medium. The player's own four values are captured first.
//   - The SESSION becomes the practice seat's, exactly as joining any
//     table swaps it. The session it replaces — a seat at a real
//     table, a Discord sign-in, the admin — is captured too.
//
// Leaving writes both back, and the ADR is explicit that "leaving"
// means EVERY exit path (§3): the Leave button, a navigation to any
// other route or site, a closed tab, a crash. A player who abandons
// the tutorial and later finds strictMana silently off in a real game
// will never connect the two. So the restore runs from four places:
//
//   1. A route change away from the practice game (the Leave button is
//      one: it navigates to the lobby).
//   2. `pagehide` — a closed tab, a reload, a navigation to another
//      site. localStorage writes are synchronous, so the settings and
//      session are back on disk before the page goes; the server half
//      rides a keepalive request.
//   3. The next page load, for the exits that run no code at all (a
//      crash, a killed process): what to restore is persisted under
//      RECORD_KEY before anything is swapped, and recoverPractice
//      finds it.
//   4. Opening a practice table while one is open ends the old one.
//
// There is no resume (§2.2): every one of those ends the practice game
// on the server too (POST /games/{id}/practice/leave), which is also
// what puts the session cookie back — the server reads the cookie
// before the Authorization header, so the client cannot restore the
// session alone (server/internal/lobby/practice_http.go).
//
// # More than one tab
//
// localStorage is shared by every tab. A practice game belongs to the
// tab that opened it (`active` below); only that tab's route changes
// and pagehide end it. It stamps the record every few seconds
// (HEARTBEAT_MS), and a page load ends a record only once that stamp
// is STALE_MS old — a second tab opened on the lobby must not end the
// first tab's tutorial. A page load whose route IS the practice game
// adopts a fresh record instead (a browser restoring a crashed tab),
// so its exit restores.

import { get, type Readable } from "svelte/store";
import { createPracticeTable, leavePracticeTable } from "./api";
import { guardedWritable } from "./guardedStore";
import { navigate, route, type Route } from "./router";
import { currentSession, savedIdentity, setSession, type Session } from "./session";
import { settings, type Settings } from "./settings";

/** The four settings the tutorial forces, as the player had them. */
export interface SavedSettings {
  strictMana: boolean;
  autoPassPriority: boolean;
  tableLayout: Settings["display"]["tableLayout"];
  cardSize: Settings["display"]["cardSize"];
}

/** The values the tutorial forces (ADR 0076 §2.2). */
export const FORCED_SETTINGS: Readonly<SavedSettings> = Object.freeze({
  strictMana: false,
  autoPassPriority: false,
  tableLayout: "quadrant",
  cardSize: "medium",
});

/**
 * What a practice game swapped out, persisted before the swap so any
 * exit — including one that runs no code — can put it back.
 */
export interface PracticeRecord {
  v: 1;
  gameID: string;
  /** The practice seat's session token: the credential for leaving. */
  practiceToken: string;
  saved: SavedSettings;
  /** The session the practice seat replaced; null if there was none. */
  previous: Session | null;
  /** Last heartbeat from the owning tab, epoch ms. */
  aliveAt: number;
}

export const RECORD_KEY = "cmdctrl.practice.v1";
/** Per-tab (sessionStorage): the practice game this tab just left. */
export const LEFT_KEY = "cmdctrl.practice.left";
export const HEARTBEAT_MS = 5_000;
export const STALE_MS = 20_000;

// --- pure helpers ---------------------------------------------------

export function captureSettings(s: Settings): SavedSettings {
  return {
    strictMana: s.gameplay.strictMana,
    autoPassPriority: s.gameplay.autoPassPriority,
    tableLayout: s.display.tableLayout,
    cardSize: s.display.cardSize,
  };
}

/**
 * withSettings returns s with the four tutorial-owned fields set to v,
 * and every other field — including anything the player changed during
 * the tutorial — untouched.
 */
export function withSettings(s: Settings, v: SavedSettings): Settings {
  return {
    ...s,
    gameplay: { ...s.gameplay, strictMana: v.strictMana, autoPassPriority: v.autoPassPriority },
    display: { ...s.display, tableLayout: v.tableLayout, cardSize: v.cardSize },
  };
}

/** True when no tab has stamped the record for STALE_MS. */
export function isStale(rec: PracticeRecord, now: number): boolean {
  return now - rec.aliveAt >= STALE_MS;
}

/** parseRecord validates a stored record; anything malformed is null. */
export function parseRecord(raw: string | null): PracticeRecord | null {
  if (!raw) return null;
  try {
    const r = JSON.parse(raw) as Partial<PracticeRecord>;
    const s = r.saved as Partial<SavedSettings> | undefined;
    if (
      r.v !== 1 ||
      typeof r.gameID !== "string" ||
      r.gameID === "" ||
      typeof r.practiceToken !== "string" ||
      typeof r.aliveAt !== "number" ||
      !s ||
      typeof s.strictMana !== "boolean" ||
      typeof s.autoPassPriority !== "boolean" ||
      (s.tableLayout !== "row" && s.tableLayout !== "quadrant") ||
      (s.cardSize !== "small" && s.cardSize !== "medium" && s.cardSize !== "large")
    ) {
      return null;
    }
    return {
      v: 1,
      gameID: r.gameID,
      practiceToken: r.practiceToken,
      saved: {
        strictMana: s.strictMana,
        autoPassPriority: s.autoPassPriority,
        tableLayout: s.tableLayout,
        cardSize: s.cardSize,
      },
      previous: r.previous ?? null,
      aliveAt: r.aliveAt,
    };
  } catch {
    return null;
  }
}

// --- storage (best-effort: private windows and full quotas throw) ----

function readRecord(): PracticeRecord | null {
  try {
    return parseRecord(localStorage.getItem(RECORD_KEY));
  } catch {
    return null;
  }
}

function writeRecord(rec: PracticeRecord): void {
  try {
    localStorage.setItem(RECORD_KEY, JSON.stringify(rec));
  } catch {
    // The in-memory copy (`active`) still restores this tab's exits.
  }
}

function clearRecord(): void {
  try {
    localStorage.removeItem(RECORD_KEY);
  } catch {
    // Nothing to do.
  }
}

function markLeft(gameID: string): void {
  try {
    sessionStorage.setItem(LEFT_KEY, gameID);
  } catch {
    // Only costs the reload redirect; see recoverPractice.
  }
}

function takeLeft(): string | null {
  try {
    const id = sessionStorage.getItem(LEFT_KEY);
    sessionStorage.removeItem(LEFT_KEY);
    return id;
  } catch {
    return null;
  }
}

// --- the practice game this tab owns ---------------------------------

let active: PracticeRecord | null = null;
let heartbeat: ReturnType<typeof setInterval> | null = null;

const store = guardedWritable<{ gameID: string } | null>(null, "practiceTable");

/**
 * practiceTable is the practice game this tab is running, or null.
 * The coach (sub-PRs 3 and 4) reads it to know the table under it is
 * the tutorial's.
 */
export const practiceTable: Readable<{ gameID: string } | null> = store;

/** isPracticeGame reports whether gameID is this tab's practice game. */
export function isPracticeGame(gameID: string): boolean {
  return active !== null && active.gameID === gameID;
}

function own(rec: PracticeRecord): void {
  active = rec;
  store.set({ gameID: rec.gameID });
  if (heartbeat !== null) clearInterval(heartbeat);
  heartbeat = setInterval(() => {
    if (active === null) return;
    active = { ...active, aliveAt: Date.now() };
    writeRecord(active);
  }, HEARTBEAT_MS);
}

function disown(): void {
  active = null;
  store.set(null);
  if (heartbeat !== null) {
    clearInterval(heartbeat);
    heartbeat = null;
  }
}

function expired(s: Session | null, now: number): boolean {
  return s === null || Date.parse(s.expiresAt) <= now;
}

/**
 * startPractice opens a practice table and swaps the session and the
 * four forced settings in, having first recorded what to put back.
 * Resolves to the table's game ID; the caller navigates there. On a
 * failed create nothing has been swapped.
 */
export async function startPractice(): Promise<string> {
  // One practice game at a time; the server would replace the old
  // table anyway, and this puts its settings and session back first.
  endPractice();

  const previous = currentSession();
  const practice = await createPracticeTable();
  if (!practice.gameID) throw new Error("the server opened no practice table");

  const rec: PracticeRecord = {
    v: 1,
    gameID: practice.gameID,
    practiceToken: practice.token,
    saved: captureSettings(get(settings)),
    previous,
    aliveAt: Date.now(),
  };
  // Recorded BEFORE anything is swapped, so an exit at any later
  // instant has what it needs.
  writeRecord(rec);
  own(rec);
  setSession(practice);
  settings.update((s) => withSettings(s, FORCED_SETTINGS));
  return rec.gameID;
}

/**
 * endPractice ends the practice game this tab is running — or, when
 * it runs none, the one on record — and writes back what it swapped
 * out. Returns false when there was nothing to end.
 *
 * Settings go back first and unconditionally. The session goes back
 * only if the tab is still holding the practice seat's: a player who
 * signed out during the tutorial stays signed out.
 */
export function endPractice(opts: { keepalive?: boolean } = {}): boolean {
  const rec = active ?? readRecord();
  if (rec === null) return false;
  const now = Date.now();

  settings.update((s) => withSettings(s, rec.saved));

  const current = currentSession();
  const holdingPractice = current !== null && current.token === rec.practiceToken;
  // The session the practice seat replaced, or, when that one has run
  // out, the signed-in session kept aside behind it (ADR 0110 §1 item
  // 6: an admin-token session that expired during the tutorial). The
  // leave call below puts the cookie back to whichever it is, so the
  // person comes out of the tutorial signed in.
  const previous = expired(rec.previous, now) ? savedIdentity(now) : rec.previous;
  // What the cookie should hold afterwards: the session this tab
  // ends up with.
  let restoreToken = current?.token ?? "";
  if (holdingPractice) {
    restoreToken = previous?.token ?? "";
  }

  clearRecord();
  disown();
  markLeft(rec.gameID);
  if (holdingPractice) setSession(previous);

  void leavePracticeTable(rec.gameID, rec.practiceToken, restoreToken, {
    keepalive: opts.keepalive,
  });
  return true;
}

/**
 * recoverPractice is the page-load half of the restore. Call it once,
 * before the app mounts:
 *
 *   - A record no tab has stamped for STALE_MS is a practice game whose
 *     tab died without running its exit. End it.
 *   - A fresh record whose game this page is opening on is a restored
 *     tab, or a duplicate of one: adopt it, so this tab's exit
 *     restores.
 *   - A fresh record for a game this page is not on belongs to another
 *     tab. Leave it alone.
 *
 * A page landing on a practice game this tab has just left (a reload:
 * pagehide ended it) is sent to the lobby rather than to a table that
 * no longer exists.
 */
export function recoverPractice(now: number = Date.now()): void {
  const left = takeLeft();
  const r: Route = get(route);
  const onGame = (id: string | null | undefined) => !!id && r.name === "game" && r.gameID === id;

  const rec = readRecord();
  let ended: string | null = null;
  if (rec !== null) {
    if (isStale(rec, now)) {
      endPractice();
      ended = rec.gameID;
    } else if (onGame(rec.gameID)) {
      own(rec);
    }
  }
  if (onGame(ended) || (onGame(left) && active === null)) {
    navigate("#/lobby");
  }
}

let installed = false;

/**
 * installPracticeExits arms the exits that need a listener: a route
 * change away from the practice game, and pagehide. Idempotent; call
 * once at boot, after recoverPractice.
 */
export function installPracticeExits(): void {
  if (installed) return;
  installed = true;
  route.subscribe((r) => {
    if (active === null) return;
    // #/practice is the door in: it is on screen between the create
    // and the navigation to the table.
    if (r.name === "practice") return;
    if (r.name === "game" && r.gameID === active.gameID) return;
    endPractice();
  });
  if (typeof window !== "undefined") {
    window.addEventListener("pagehide", () => {
      if (active !== null) endPractice({ keepalive: true });
    });
  }
}
