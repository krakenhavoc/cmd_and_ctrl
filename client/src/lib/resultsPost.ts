// resultsPost.ts — a table's opt-in to a results post in the CMD & CTRL
// Discord's #game-results channel (ADR 0174): the one route, and what
// the table settings say about it.
//
// The server decides everything; this file only words it. The choice
// arrives on the table's meta (`results`) from GET /games and from the
// PUT's answer.

import type { GameMeta } from "./api";
import { authFetch } from "./session";

// setResults opts the table in to a results post, or out of it (PUT
// /games/{id}/results). Host or admin; the server refuses anyone else,
// and says why a table cannot be opted in. Answers the table's meta.
export async function setResults(gameID: string, enabled: boolean): Promise<GameMeta> {
  const res = await authFetch(`/games/${encodeURIComponent(gameID)}/results`, {
    method: "PUT",
    body: JSON.stringify({ enabled }),
  });
  return (await res.json()) as GameMeta;
}

// resultsBlocked is why the box cannot be ticked for g, or null when it
// can. Opting out is always allowed, so a table already opted in is
// never blocked. The server checks the same things.
export function resultsBlocked(g: GameMeta): string | null {
  if (g.results) return null;
  if (g.practice) return "A practice table is never posted.";
  if (g.state === "ended" || g.archived_at) {
    return "This game has ended, so its result can no longer be posted.";
  }
  return null;
}

// resultsStatusLine is the panel's one-sentence account of the choice.
// The post is made once, when the game ends, so there is no status to
// report beyond whether it is still to come.
export function resultsStatusLine(g: GameMeta): string {
  if (!g.results) return "This table's result will not be posted to Discord.";
  if (g.state === "ended")
    return "This game has ended. Its result was posted if it was played out.";
  return "The result will be posted to Discord when the game ends. A game conceded in its first turns is not posted.";
}
