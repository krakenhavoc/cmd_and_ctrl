import type { Session } from "./session";

// myGames.ts is the pure half of "My games" (ADR 0051 decision 4, S34
// sub-PR 4): the wire shape of GET /me/games, who counts as signed in,
// when the in-game "Link Discord" entry is offered, and the labels the
// MyGames route renders. No DOM and no fetch, so all of it is unit-
// tested in the node environment (myGames.test.ts).

// MySeat mirrors lobby.MySeat: another seat at the table.
export interface MySeat {
  seat: number;
  name: string;
  bot?: boolean;
}

// MyGame mirrors lobby.MyGame, one entry of GET /me/games. Every time
// is Unix MILLISECONDS (the games table's unit), and a time that has
// not happened is null. `rejoin` is present only while the table is
// still open; it is the path to POST for a fresh seat session.
export interface MyGame {
  id: string;
  name: string;
  state: "lobby" | "active" | "ended";
  seat: number;
  winner_seat: number | null;
  created_at: number;
  started_at: number | null;
  ended_at: number | null;
  archived_at: number | null;
  others: MySeat[];
  rejoin?: string;
}

// The server spells "no user" as the nil uuid on a guest's principal
// (uuid.UUID has no omitempty), and leaves it out entirely on a
// deployment with no database. Both mean the same thing.
const NIL_UUID = "00000000-0000-0000-0000-000000000000";

// signedInUserID returns the session's user id when it belongs to a
// signed-in person: an identity session, or a player session minted
// from one. Admin, spectator and guest sessions return null, and so
// does a missing session. This is the gate for every "My games" link.
export function signedInUserID(s: Session | null | undefined): string | null {
  if (!s) return null;
  const { role, user_id: id } = s.principal;
  if (!id || id === NIL_UUID) return null;
  if (role !== "identified" && role !== "player") return null;
  return id;
}

// canLinkDiscord decides whether the in-game menu offers "Link
// Discord": a seated human player on a server with Discord configured.
// A guest's seat is the case the feature exists for (ADR 0051: a guest
// at a live table can sign in and become that seat's user); a Discord
// seat may relink to another account. Admins, spectators and bot seats
// have no seat of their own to link.
export function canLinkDiscord(opts: {
  role: Session["principal"]["role"] | undefined;
  discordEnabled: boolean;
  isBotSeat: boolean;
}): boolean {
  return opts.discordEnabled && opts.role === "player" && !opts.isBotSeat;
}

// linkDiscordLabel is the menu entry's text: "Link Discord" for a guest
// seat, and a relink for a seat that already carries an account.
export function linkDiscordLabel(seatHasDiscord: boolean): string {
  return seatHasDiscord ? "Link a different Discord account" : "Link Discord";
}

export type MyGameTone = "lobby" | "live" | "won" | "ended" | "archived";

// myGameStatus is the chip on one row. Archived wins over the state:
// the table is retired whatever it was doing.
export function myGameStatus(g: MyGame): { label: string; tone: MyGameTone } {
  if (g.archived_at !== null) return { label: "archived", tone: "archived" };
  switch (g.state) {
    case "lobby":
      return { label: "in the lobby", tone: "lobby" };
    case "active":
      return { label: "in progress", tone: "live" };
    default: {
      if (g.winner_seat === null) return { label: "ended", tone: "ended" };
      if (g.winner_seat === g.seat) return { label: "you won", tone: "won" };
      const winner = g.others.find((o) => o.seat === g.winner_seat);
      return { label: winner ? `${winner.name} won` : "ended", tone: "ended" };
    }
  }
}

// othersLabel names the rest of the table in seat order: "with Bob,
// Carol and Ghoul (bot)". A table nobody else sat at says so.
export function othersLabel(g: MyGame): string {
  const names = [...g.others]
    .sort((a, b) => a.seat - b.seat)
    .map((o) => (o.bot ? `${o.name} (bot)` : o.name))
    .filter((n) => n.trim() !== "");
  if (names.length === 0) return "nobody else yet";
  if (names.length === 1) return `with ${names[0]}`;
  return `with ${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}`;
}

const DAY_MS = 24 * 60 * 60 * 1000;

// playedWhen is the row's date: the most recent thing that happened to
// the table (ended, else started, else created), relative while it is
// recent and a plain date after a week.
export function playedWhen(g: MyGame, now: number = Date.now()): string {
  const at = g.ended_at ?? g.started_at ?? g.created_at;
  const startOfToday = new Date(now);
  startOfToday.setHours(0, 0, 0, 0);
  const today = startOfToday.getTime();
  if (at >= today) return "today";
  if (at >= today - DAY_MS) return "yesterday";
  const days = Math.ceil((today - at) / DAY_MS);
  if (days < 7) return `${days} days ago`;
  return new Date(at).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// sortMyGames is newest first by creation, the order the server
// already sends, re-applied so the page never depends on it.
export function sortMyGames(games: MyGame[]): MyGame[] {
  return [...games].sort((a, b) => b.created_at - a.created_at || a.id.localeCompare(b.id));
}
