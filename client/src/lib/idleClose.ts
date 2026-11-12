// idleClose.ts: what the client says about ADR 0211's idle tables.
//
//   - The table's notice (§6). Thirty-six hours after the last move by a
//     person, the server warns a started game that it will close in
//     twelve, and every frame carries `idle_close_at` until somebody
//     moves. The table shows one line over the action dock:
//     "This game closes at 21:14 on Friday unless someone moves", in the
//     viewer's own clock.
//   - The close reason (owner decision 4). A table the server closed
//     carries `closed_reason` in GET /me/games and the admin views, and
//     its result chip says why instead of a bare "Archived".
//
// Pure: no DOM and no fetch, so idleClose.test.ts covers all of it.
import { L } from "./labels";

const DAY_MS = 24 * 60 * 60 * 1000;

function startOfDay(ms: number): number {
  const d = new Date(ms);
  d.setHours(0, 0, 0, 0);
  return d.getTime();
}

/**
 * idleCloseWhen is the closing time as the notice reads it, in the
 * viewer's local time: "at 21:14 today", "at 09:30 tomorrow", or "at
 * 21:14 on Friday" further out (the warning comes twelve hours ahead, so
 * a weekday is as far as it ever needs to go).
 */
export function idleCloseWhen(closesAt: number, now: number = Date.now()): string {
  const time = new Date(closesAt).toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
  const days = Math.round((startOfDay(closesAt) - startOfDay(now)) / DAY_MS);
  if (days <= 0) return `at ${time} today`;
  if (days === 1) return `at ${time} tomorrow`;
  const weekday = new Date(closesAt).toLocaleDateString(undefined, { weekday: "long" });
  return `at ${time} on ${weekday}`;
}

/**
 * idleCloseNotice is the table's one line for a frame's `idle_close_at`,
 * or null when the game is not warned (or the value does not parse).
 */
export function idleCloseNotice(
  idleCloseAt: string | undefined,
  now: number = Date.now(),
): string | null {
  if (!idleCloseAt) return null;
  const at = Date.parse(idleCloseAt);
  if (!Number.isFinite(at)) return null;
  return L.idleCloseNotice(idleCloseWhen(at, now));
}

/**
 * closedReasonLabel is the result chip for a table the server closed for
 * being idle (`closed_reason`: idle_lobby, idle_game or idle_bots), or
 * null for anything else; an admin's archive has no reason.
 */
export function closedReasonLabel(reason: string | null | undefined): string | null {
  switch (reason) {
    case "idle_lobby":
      return L.closedIdleLobby;
    case "idle_game":
      return L.closedIdleGame;
    case "idle_bots":
      return L.closedIdleBots;
  }
  return null;
}
