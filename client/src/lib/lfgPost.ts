// lfgPost.ts — a table's post on the CMD & CTRL Discord's #lfg board
// (ADR 0161): the one route, and what the table settings say about it.
//
// The server decides everything; this file only words it. The post's
// state arrives on the table's meta (`lfg`, `lfg_status`) from GET
// /games and from the PUT's answer.

import type { GameMeta } from "./api";
import { authFetch } from "./session";

/** What the server says about a table's post (GameMeta.lfg_status). */
export type LFGStatus = "pending" | "posted" | "unavailable";

/** The hint a host who cannot post sees under the disabled box. */
export const LFG_GUEST_HINT = "Sign in with Discord to post to #lfg";

// setLFG posts the table to #lfg, or withdraws it (PUT /games/{id}/lfg).
// Host or admin; the server refuses anyone else, and says why a table
// cannot be posted. Answers the table's meta.
export async function setLFG(gameID: string, enabled: boolean): Promise<GameMeta> {
  const res = await authFetch(`/games/${encodeURIComponent(gameID)}/lfg`, {
    method: "PUT",
    body: JSON.stringify({ enabled }),
  });
  return (await res.json()) as GameMeta;
}

// lfgBlocked is why the box cannot be ticked for g, or null when it can.
// Turning a post off is always allowed, so a table already posted is
// never blocked. The server checks the same things.
export function lfgBlocked(g: GameMeta): string | null {
  if (g.lfg) return null;
  if (g.state !== "lobby" || g.archived_at) return "Only a table waiting to start can be posted.";
  if (g.players.length >= 4) return "Every seat is taken.";
  const host = g.players.find((p) => p.is_host);
  if (host && !host.discord_id) return LFG_GUEST_HINT;
  return null;
}

/** What the table card's header shows for the #lfg post. */
export type LFGHeader =
  | { kind: "post"; disabled: boolean; title: string }
  | { kind: "posted"; label: string; title: string };

// lfgHeader is the table card's #lfg control, beside Copy invite, for a
// host or admin of a table waiting to start; null when there is none.
// Unposted, it is a button that posts. Posted, it is a status, and the
// ⋯ menu's "Withdraw from #lfg" takes the post down.
export function lfgHeader(g: GameMeta): LFGHeader | null {
  if (g.state !== "lobby" || g.archived_at) return null;
  if (!g.lfg) {
    const blocked = lfgBlocked(g);
    return {
      kind: "post",
      disabled: blocked !== null,
      title: blocked ?? "post this table to the CMD & CTRL Discord's #lfg channel and ping LFG",
    };
  }
  switch (g.lfg_status) {
    case "posted":
      return {
        kind: "posted",
        label: "In #lfg",
        title: "posted in #lfg; it comes down when the game starts",
      };
    case "unavailable":
      return {
        kind: "posted",
        label: "#lfg unavailable",
        title: "Discord refused the post or didn't answer; withdraw and post again to retry",
      };
    default:
      return { kind: "posted", label: "Posting…", title: "posting to #lfg" };
  }
}

// lfgStatusLine is the panel's one-sentence account of the post, or
// null when the table is not posted.
export function lfgStatusLine(g: GameMeta): string | null {
  if (!g.lfg) return null;
  if (g.state !== "lobby" || g.archived_at) return "The #lfg post comes down once the game starts.";
  switch (g.lfg_status) {
    case "posted":
      return "Posted in #lfg. The post follows the seats and comes down when the game starts.";
    case "unavailable":
      return "The #lfg post is unavailable: Discord refused it or didn't answer. The table is unaffected; turn it off and on to try again.";
    default:
      return "Posting to #lfg…";
  }
}
