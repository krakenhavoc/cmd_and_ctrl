// endedGame.ts — whether the signed-in person has a finished game, for
// the tutorial offer (ADR 0125 §3.7 `lobby.practice`, §6).
//
// The offer goes to a guest, or to a signed-in person whose GET
// /me/games has no ended game. Nothing new is fetched for it: the pages
// that already read /me/games (the Lobby's "your tables", My games) hand
// what they got to `noteMyGames`, and HintLayer reads the answer back
// through `hasEndedGameFor` into `HintContext.hasEndedGame`.
//
// The answer is kept per user, so a different sign-in on the same tab
// starts from "unknown" again, and it is null until a read has landed:
// the hint waits for it rather than guessing (LobbyPractice.hint.ts).

import { get } from "svelte/store";
import { guardedWritable } from "../guardedStore";
import { signedInUserID, type MyGame } from "../myGames";
import type { Session } from "../session";

interface Known {
  userID: string;
  ended: boolean;
}

const known = guardedWritable<Known | null>(null, "hasEndedGame");

/** anyEnded reports whether a /me/games list holds a finished game. */
export function anyEnded(games: readonly Pick<MyGame, "state">[]): boolean {
  return games.some((g) => g.state === "ended");
}

/** noteMyGames records what a GET /me/games for this user said. */
export function noteMyGames(userID: string | null, games: readonly Pick<MyGame, "state">[]): void {
  if (!userID) return;
  known.set({ userID, ended: anyEnded(games) });
}

/**
 * hasEndedGameFor is whether this session's person has a finished game:
 * true or false once a read of /me/games has landed for them, null while
 * unknown and for anyone with no user (a guest, the admin token).
 */
export function hasEndedGameFor(s: Session | null | undefined): boolean | null {
  const id = signedInUserID(s);
  const k = get(known);
  return id !== null && k !== null && k.userID === id ? k.ended : null;
}

/** _resetEndedGameForTests forgets every answer. */
export function _resetEndedGameForTests(): void {
  known.set(null);
}
