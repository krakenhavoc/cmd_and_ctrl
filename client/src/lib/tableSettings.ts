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

import type { PlayerView } from "./protocol";

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
