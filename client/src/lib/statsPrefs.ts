// statsPrefs.ts — the advanced-stats setting (ADR 0217 §2, owner
// answer 1; slice 3 of #4880): the person's own switch behind GET and
// PUT /me/stats/prefs, the store Settings and #/stats share, and the
// admin side (GET /admin/stats/advanced, PUT /admin/stats/advanced-default,
// PUT /admin/users/{id}/advanced-stats).
//
// The server decides everything: a person's own answer, else an admin's
// switch for them, else the server's default, which is off unless an
// admin stores one. This file only reads and writes. A guest is never
// asked: guests get no stats.

import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { signedInUserID } from "./myGames";
import { authFetch, LobbyApiError, session } from "./session";

/** "on" or "off" when someone chose; "default" when it follows the server's. */
export type StatsChoice = "on" | "off" | "default";
export type StatsSetBy = "self" | "admin";

/** The body of GET and PUT /me/stats/prefs. */
export interface StatsPrefs {
  /** False with no database, or for a caller who is not a signed-in person. */
  available: boolean;
  /** The answer now. */
  advanced: boolean;
  choice: StatsChoice;
  set_by?: StatsSetBy;
  /** The server's default, which "default" follows. */
  default: boolean;
  /**
   * When the person's share link was made, Unix ms (ADR 0217 §6); absent
   * while they have none. The link itself is shown once, when made.
   */
  share_link_at?: number;
  /**
   * Whether the person shares their stats with the members of every pod
   * they are in (ADR 0220 §7): one switch for all pods, off by default.
   */
  share_pods: boolean;
}

/** What every surface shows until the server answers: off, nothing to switch. */
export const STATS_PREFS_OFF: StatsPrefs = Object.freeze({
  available: false,
  advanced: false,
  choice: "default",
  default: false,
  share_pods: false,
});

function normalizeChoice(c: unknown): StatsChoice {
  return c === "on" || c === "off" ? c : "default";
}

function normalizeSetBy(v: unknown): StatsSetBy | undefined {
  return v === "self" || v === "admin" ? v : undefined;
}

export function normalizeStatsPrefs(raw: unknown): StatsPrefs {
  const r = (raw ?? {}) as Partial<StatsPrefs>;
  const out: StatsPrefs = {
    available: r.available === true,
    advanced: r.advanced === true,
    choice: normalizeChoice(r.choice),
    default: r.default === true,
    share_pods: r.share_pods === true,
  };
  const by = normalizeSetBy(r.set_by);
  if (by && out.choice !== "default") out.set_by = by;
  if (typeof r.share_link_at === "number" && r.share_link_at > 0)
    out.share_link_at = r.share_link_at;
  return out;
}

/** The setting as the server last said, off until it has. */
export const statsPrefs: Writable<StatsPrefs> = guardedWritable(STATS_PREFS_OFF, "statsPrefs");

/**
 * loadStatsPrefs reads the setting into the store. A caller that is not
 * a signed-in person is not asked (the server would answer 403); any
 * refusal leaves the store off, which hides the switch.
 */
export async function loadStatsPrefs(): Promise<StatsPrefs> {
  if (signedInUserID(get(session)) === null) {
    statsPrefs.set(STATS_PREFS_OFF);
    return STATS_PREFS_OFF;
  }
  try {
    const res = await authFetch("/me/stats/prefs", { cache: "no-store" });
    const got = normalizeStatsPrefs(await res.json());
    statsPrefs.set(got);
    return got;
  } catch (e) {
    if (!(e instanceof LobbyApiError && e.status === 403)) {
      console.warn("could not read the stats setting", e);
    }
    statsPrefs.set(STATS_PREFS_OFF);
    return STATS_PREFS_OFF;
  }
}

/**
 * setAdvancedStats switches advanced stats on or off for the caller. It
 * resolves to the server's answer, already in the store, and rejects
 * with the server's message when the write is refused.
 */
export async function setAdvancedStats(on: boolean): Promise<StatsPrefs> {
  const res = await authFetch("/me/stats/prefs", {
    method: "PUT",
    body: JSON.stringify({ advanced: on }),
  });
  const got = normalizeStatsPrefs(await res.json());
  statsPrefs.set(got);
  return got;
}

/**
 * setSharePods switches pod sharing on or off for the caller (ADR 0220
 * §7). It leaves the advanced-stats switch alone; the server's answer is
 * already in the store when it resolves.
 */
export async function setSharePods(on: boolean): Promise<StatsPrefs> {
  const res = await authFetch("/me/stats/prefs", {
    method: "PUT",
    body: JSON.stringify({ share_pods: on }),
  });
  const got = normalizeStatsPrefs(await res.json());
  statsPrefs.set(got);
  return got;
}

/** The line under the person's switch: why it is in the state it is. */
export function statsPrefsNote(p: StatsPrefs): string {
  if (p.choice === "default") {
    return p.default
      ? "On for everyone who hasn't chosen. Switch it off if you'd rather not see them."
      : "Off until you turn it on.";
  }
  if (p.set_by === "admin") {
    return p.advanced
      ? "An admin turned this on for you. You can switch it off."
      : "An admin turned this off for you. You can switch it on.";
  }
  return p.advanced ? "You turned this on." : "You turned this off.";
}

// ---- The admin side ----

/** One account's line in GET /admin/stats/advanced. */
export interface AdminAccountStats {
  user_id: string;
  choice: StatsChoice;
  set_by?: StatsSetBy;
  advanced: boolean;
  updated_at?: number;
}

/** The server's default, as the admin routes answer it. */
export interface AdminStatsDefault {
  on: boolean;
  /** False while no admin has set one; `on` is then the code's fallback (off). */
  stored: boolean;
  updated_at?: number;
}

/** The body of GET /admin/stats/advanced. */
export interface AdminAdvancedStats {
  default: AdminStatsDefault;
  /** Every account whose switch is set. An account not listed follows the default. */
  accounts: AdminAccountStats[];
  truncated?: boolean;
}

export async function fetchAdminAdvancedStats(): Promise<AdminAdvancedStats> {
  const res = await authFetch("/admin/stats/advanced", { cache: "no-store" });
  const raw = (await res.json()) as Partial<AdminAdvancedStats>;
  return {
    default: {
      on: raw.default?.on === true,
      stored: raw.default?.stored === true,
      updated_at: raw.default?.updated_at,
    },
    accounts: Array.isArray(raw.accounts) ? raw.accounts : [],
    truncated: raw.truncated === true,
  };
}

export async function setAdminAdvancedDefault(on: boolean): Promise<AdminStatsDefault> {
  const res = await authFetch("/admin/stats/advanced-default", {
    method: "PUT",
    body: JSON.stringify({ on }),
  });
  return (await res.json()) as AdminStatsDefault;
}

export async function setAccountAdvancedStats(
  userID: string,
  on: boolean,
): Promise<AdminAccountStats> {
  const res = await authFetch(`/admin/users/${encodeURIComponent(userID)}/advanced-stats`, {
    method: "PUT",
    body: JSON.stringify({ on }),
  });
  return (await res.json()) as AdminAccountStats;
}

/**
 * accountStatsOf is one account's state for the admin listing: its row
 * when its switch is set, else following the default.
 */
export function accountStatsOf(data: AdminAdvancedStats, userID: string): AdminAccountStats {
  return (
    data.accounts.find((a) => a.user_id === userID) ?? {
      user_id: userID,
      choice: "default",
      advanced: data.default.on,
    }
  );
}

/** withAccount replaces (or adds) one account's line after a switch. */
export function withAccount(data: AdminAdvancedStats, row: AdminAccountStats): AdminAdvancedStats {
  return {
    ...data,
    accounts: [row, ...data.accounts.filter((a) => a.user_id !== row.user_id)],
  };
}

/**
 * withDefault is the listing after the default changed: the default
 * itself, and the answer of every listed row that follows it (none do
 * today, but a later slice's row may exist with no switch set).
 */
export function withDefault(data: AdminAdvancedStats, def: AdminStatsDefault): AdminAdvancedStats {
  return {
    ...data,
    default: def,
    accounts: data.accounts.map((a) => (a.choice === "default" ? { ...a, advanced: def.on } : a)),
  };
}

/** The admin's words for one account's state: on, off, or following the default. */
export function accountStatsText(a: AdminAccountStats): string {
  if (a.choice === "default") return `Default (${a.advanced ? "on" : "off"})`;
  const state = a.choice === "on" ? "On" : "Off";
  return a.set_by === "admin" ? `${state}, set by an admin` : `${state}, set by them`;
}

/** shareLinkChanged keeps the store in step after the link is made or revoked. */
export function shareLinkChanged(at: number | null): void {
  statsPrefs.update((p) => {
    const next: StatsPrefs = { ...p };
    if (at && at > 0) next.share_link_at = at;
    else delete next.share_link_at;
    return next;
  });
}
