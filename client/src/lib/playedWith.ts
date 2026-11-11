// playedWith.ts — the Played with list in the header's drawer (ADR 0213
// §1, slice 2): which of your tables an invite can go to, and the words
// each row and each choice says.
//
// Pure, like tablemates.ts and tablesPage.ts: the decisions are worth a
// test and a `.svelte` file is the awkward place to write one. The drawer
// (components/PlayedWithDrawer.svelte) only fetches and draws.

import type { GameMeta } from "./api";
import type { Session } from "./session";
import { canInviteTablemates, type Tablemate } from "./tablemates";
import { tableSeats } from "./formats";

/**
 * openSeats is how many seats a table has free. A table seats four
 * (ADR 0212 owner decision 5); bots hold seats too, so they count as
 * taken.
 */
export function openSeats(g: GameMeta): number {
  return Math.max(0, tableSeats(g) - g.players.length);
}

/**
 * inviteTables are the tables an invite from the Played with list can
 * name (ADR 0213 §1): your own tables still in the lobby with an empty
 * seat. "Your own" is the rule the DM route holds the caller to
 * (canInviteTablemates: seated there, or its creator), so a table the
 * server would refuse is never offered. A practice table is never in
 * the list GET /games serves, and is dropped here all the same.
 */
export function inviteTables(games: GameMeta[], s: Session | null | undefined): GameMeta[] {
  return games.filter(
    (g) =>
      g.state === "lobby" &&
      g.practice !== true &&
      openSeats(g) > 0 &&
      canInviteTablemates(s, s?.gameID, g.id, g.is_creator === true),
  );
}

/** tableChoiceLabel is one table in the chooser: its name and how full it is. */
export function tableChoiceLabel(g: GameMeta): string {
  return `${g.name} · ${g.players.length}/${tableSeats(g)}`;
}

function plural(n: number, one: string, many: string): string {
  return `${n} ${n === 1 ? one : many}`;
}

/**
 * togetherLine is the line under a name in the Played with list: how
 * many tables you have shared and the latest one's name, then how long
 * ago it was. A server that predates the two fields (games_together,
 * last_table) still gets a sensible line.
 *
 * `now` is injected so a test does not depend on the clock.
 */
export function togetherLine(mate: Tablemate, now: number = Date.now()): string {
  const count = mate.games_together ?? 0;
  const parts: string[] = [];
  if (count > 0) parts.push(plural(count, "table together", "tables together"));
  if (mate.last_table) parts.push(`last: ${mate.last_table}`);
  const ago = agoLabel(mate.last_played_at, now);
  if (ago) parts.push(ago);
  return parts.join(" · ") || "played together";
}

/** agoLabel is "today", "yesterday", "3 days ago" … for a Unix-millisecond time. */
export function agoLabel(at: number | undefined, now: number): string {
  if (!at || at <= 0) return "";
  const days = Math.floor((now - at) / 86_400_000);
  if (days <= 0) return "today";
  if (days === 1) return "yesterday";
  if (days < 30) return `${days} days ago`;
  const months = Math.floor(days / 30);
  if (months < 12) return months === 1 ? "a month ago" : `${months} months ago`;
  const years = Math.floor(days / 365);
  return years <= 1 ? "a year ago" : `${years} years ago`;
}

/**
 * inviteHeading is what the open chooser asks. The delivery depends on
 * the server: with a bot token it is a Discord DM, without one (every
 * environment but production) it is a link to copy.
 */
export function inviteHeading(name: string, dm: boolean): string {
  return dm ? `Send ${name} an invite by Discord DM to` : `Copy an invite link for ${name} to`;
}

/** inviteDone is the line shown once an invite has gone out, or been copied. */
export function inviteDone(name: string, table: string, dm: boolean): string {
  return dm ? `Invite to ${table} sent to ${name}` : `Invite link for ${table} copied for ${name}`;
}
