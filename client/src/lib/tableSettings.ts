// tableSettings.ts — the client's half of ADR 0075 §2.2/§2.3: who may
// change a table's settings, and how an unlimited undo budget reads.
//
// Both answers are one-liners, and both are here rather than inline in
// Game.svelte for the same reason: they are asked in several places
// (the menu row, the Undo button's disabled state, the keyboard
// shortcut's gate) and a copy that disagrees with the others is a
// control that lies about what the server will accept.
//
// The full settings panel is ADR 0075 sub-PR 5. This file carries only
// what the in-game menu needs today.

import type { GameView, PlayerView, TableSettingsView } from "./protocol";

// The roles a session can carry (session.ts `Session.principal.role`).
type PrincipalRole = "player" | "admin" | "spectator" | "identified";

// UNDO_UNLIMITED mirrors game.UndoUnlimited. It is a real value, not a
// sentinel for "unset": a table can deliberately run with no budget,
// and the server never rewrites it.
export const UNDO_UNLIMITED = -1;

/**
 * True when the budget is unlimited — the server debits nothing and
 * refuses nothing.
 *
 * The negative test matters more than it looks. Under an unlimited
 * budget every seat's `undos_remaining` is -1, so any check shaped
 * `remaining <= 0` greys the Undo button on precisely the table that
 * has the most undos. Ask this first.
 */
export function isUnlimitedUndo(n: number | undefined | null): boolean {
  return n !== undefined && n !== null && n <= UNDO_UNLIMITED;
}

/** Renders an undo count for a human: "∞" for unlimited, else the number. */
export function formatUndoCount(n: number | undefined | null): string {
  if (isUnlimitedUndo(n)) return "∞";
  return String(n ?? 0);
}

/**
 * True when the viewer may spend an undo right now: an admin (who
 * bypasses the budget entirely), an unlimited table, or a seat with
 * budget left.
 */
export function hasUndoBudget(seat: PlayerView | null, isAdmin: boolean): boolean {
  if (isAdmin) return true;
  const n = seat?.undos_remaining ?? 0;
  return isUnlimitedUndo(n) || n > 0;
}

/**
 * True when the viewer may change the table's settings — the table
 * host, or the server admin (ADR 0075 §2.1, `lobby.CanManageTable`).
 *
 * This is a mirror of a server-side gate, not the gate itself: the
 * server refuses a non-manager whatever the client renders. It exists
 * so the control is greyed rather than offering a click that comes
 * back as an error frame.
 */
export function canManageTable(role: PrincipalRole | undefined, seat: PlayerView | null): boolean {
  if (role === "admin") return true;
  // The role check is not redundant with the seat check. A spectator
  // never matches a seat of its own today, so `viewerSeat` is null
  // for one — but that is a property of how the caller finds the
  // seat, not a rule, and the server's predicate names the role
  // explicitly (auth.RolePlayer). Mirror it.
  return role === "player" && Boolean(seat?.is_host);
}

// --- the settings panel (ADR 0075 §2.5) -----------------------------
//
// Everything below is what the panel needs to render and what it is
// allowed to send. It lives here rather than in the component because
// the panel has two hosts — the lobby page, which PATCHes over HTTP,
// and the in-game menu, which sends the `set_table_settings` action
// over the socket — and both must agree about ranges, labels and
// which control is locked. A rule written twice is a rule that will
// eventually be written differently.

// A partial update. The wire body IS this object: only the fields
// present are applied, which is the difference between turning
// spawning on and resetting the undo limit as a side effect.
export type TableSettingsPatch = Partial<TableSettingsView>;

// The server's ranges (game/settings.go). Mirrored so a control can
// refuse a value before it becomes a 400 the player has to read.
export const MIN_STARTING_LIFE = 1;
export const MAX_STARTING_LIFE = 999;
export const MIN_COMMANDER_DAMAGE = 1;
export const MAX_COMMANDER_DAMAGE = 99;
// The engine accepts any budget from UNDO_UNLIMITED up. The upper
// bound is the client's own: past a couple of dozen take-backs per
// turn the number has stopped meaning anything and "unlimited" is the
// honest setting.
export const MAX_UNDO_LIMIT = 20;

// DEFAULT_TABLE_SETTINGS mirrors game.DefaultTableSettings. Used when
// a snapshot carries no `settings` at all — an older server, or a
// frame captured before S35 — so the panel renders the truth about
// what that table is doing rather than a row of blanks.
export const DEFAULT_TABLE_SETTINGS: TableSettingsView = {
  undo_limit: 1,
  undo_scope: "own",
  starting_life: 40,
  commander_damage: 21,
  bot_pace: "normal",
  allow_spawn: false,
};

/** The table's settings, or the defaults when the frame has none. */
export function tableSettingsOf(view: GameView | null | undefined): TableSettingsView {
  return view?.settings ?? DEFAULT_TABLE_SETTINGS;
}

// A choice a segmented control or a select offers: the wire value,
// the words on the control, and the sentence under it. The hint is
// written for a player, not for an engineer — "the host can take back
// anyone's move", never "UndoScopeHostAny".
export interface SettingChoice<T extends string> {
  value: T;
  label: string;
  hint: string;
}

export const UNDO_SCOPE_CHOICES: SettingChoice<TableSettingsView["undo_scope"]>[] = [
  {
    value: "own",
    label: "Their own",
    hint: "Each player can only take back their own moves.",
  },
  {
    value: "host_any",
    label: "Host can undo anyone",
    hint: "The host can also take back somebody else's move — useful when one seat has wandered off.",
  },
];

export const BOT_PACE_CHOICES: SettingChoice<TableSettingsView["bot_pace"]>[] = [
  { value: "fast", label: "Fast", hint: "Bots answer almost immediately. Good for testing." },
  { value: "normal", label: "Normal", hint: "A beat of thought before each move, like a person." },
  { value: "slow", label: "Slow", hint: "Bots take their time, so the table can follow along." },
];

/**
 * The undo budget in words, for the line under the control. The
 * numbers do not read as sentences on their own: -1 is not "minus one
 * take-back" and 0 is not "no limit".
 */
export function describeUndoLimit(n: number): string {
  if (isUnlimitedUndo(n)) return "Unlimited take-backs. Nothing is ever refused.";
  if (n === 0) return "Undo is off. Nobody can take a move back.";
  if (n === 1) return "One take-back per player, per turn.";
  return `${n} take-backs per player, per turn.`;
}

/**
 * True once starting life is fixed (ADR 0075 §2.3). The value has
 * already been dealt out, and rewriting life totals mid-game would be
 * a different, sneakier feature — the server answers 422, so the
 * control is disabled rather than offering the click.
 *
 * Anything that is not the lobby counts, an ended game included: a
 * finished table has nothing to deal.
 */
export function startingLifeLocked(state: string | undefined | null): boolean {
  return state !== "lobby";
}

/**
 * The reason a patch cannot be sent, or null when it can. Mirrors
 * game.SettingsPatch.Validate plus the one state rule the engine adds
 * on top; the server is still the authority, and this exists so an
 * out-of-range number is caught under the control that produced it.
 */
export function settingsPatchError(
  patch: TableSettingsPatch,
  state?: string | null,
): string | null {
  if (patch.undo_limit !== undefined) {
    const n = patch.undo_limit;
    if (!Number.isInteger(n) || n < UNDO_UNLIMITED) {
      return `An undo limit is ${UNDO_UNLIMITED} for unlimited, or 0 and up.`;
    }
  }
  if (patch.starting_life !== undefined) {
    const n = patch.starting_life;
    if (!Number.isInteger(n) || n < MIN_STARTING_LIFE || n > MAX_STARTING_LIFE) {
      return `Starting life is ${MIN_STARTING_LIFE}–${MAX_STARTING_LIFE}.`;
    }
    if (startingLifeLocked(state)) {
      return "The game has already started, so starting life is fixed. Use the life controls on each seat.";
    }
  }
  if (patch.commander_damage !== undefined) {
    const n = patch.commander_damage;
    if (!Number.isInteger(n) || n < MIN_COMMANDER_DAMAGE || n > MAX_COMMANDER_DAMAGE) {
      return `Commander damage is ${MIN_COMMANDER_DAMAGE}–${MAX_COMMANDER_DAMAGE}.`;
    }
  }
  return null;
}

/**
 * The fields of `next` that differ from `current`. The panel applies
 * a field at a time, so this is mostly a guard against sending a
 * patch for a control the user opened and closed again: a no-op patch
 * is accepted by the server but still commits a frame and wakes every
 * connected client.
 */
export function settingsDiff(
  current: TableSettingsView,
  next: TableSettingsPatch,
): TableSettingsPatch {
  const out: TableSettingsPatch = {};
  for (const key of Object.keys(next) as (keyof TableSettingsView)[]) {
    const v = next[key];
    if (v !== undefined && v !== current[key]) {
      // The cast is the price of iterating a union of six field
      // types; every value came out of `next` under the same key.
      (out as Record<string, unknown>)[key] = v;
    }
  }
  return out;
}

/**
 * True when the viewer may open the spawner: BOTH gates the server
 * checks (ADR 0075 §2.4). Offering it on one of them produces a
 * button whose 403 explains a rule the player could have been shown
 * instead.
 */
export function canSpawn(
  role: PrincipalRole | undefined,
  seat: PlayerView | null,
  settings: TableSettingsView,
): boolean {
  return canManageTable(role, seat) && settings.allow_spawn;
}

/**
 * True while the table should be wearing the "spawning on" badge.
 *
 * Deliberately not gated on who is looking. The badge exists FOR the
 * opponents: a Treasure that came from nowhere is indistinguishable
 * from a real one, and the table's answer to that is that everybody
 * can see the switch is on and every use is in the log.
 */
export function spawningVisible(settings: TableSettingsView): boolean {
  return settings.allow_spawn;
}

// --- the spawner (ADR 0075 §2.4) ------------------------------------

export type SpawnZone = "battlefield" | "hand" | "graveyard" | "exile" | "library" | "command";

export const SPAWN_ZONES: SpawnZone[] = [
  "battlefield",
  "hand",
  "graveyard",
  "exile",
  "library",
  "command",
];

/**
 * The zones a spawn of this kind may target.
 *
 * A token may only go to the battlefield: CR 704.5d removes it from
 * every other zone at the next state-based action check, so spawning
 * one into a hand appears to work and then silently undoes itself.
 * The server refuses it (400); this keeps the option off the menu.
 */
export function spawnZonesFor(kind: "card" | "token"): SpawnZone[] {
  return kind === "token" ? ["battlefield"] : SPAWN_ZONES;
}
