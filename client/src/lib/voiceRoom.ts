// voiceRoom.ts — a table's Discord voice room (ADR 0159 §6): the two
// routes, and what the game bar and the voice panel say about it.
//
// The server decides everything; this file only words it. The room's
// state arrives on the game view (`view.voice`), already reduced to this
// viewer's own access, with the link only when they were let in.

import type { Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { authFetch } from "./session";
import type { VoiceConn, VoiceView } from "./protocol";

// voiceSeats is who is connected to the table's voice room, by player
// id (#3311). Game.svelte keeps it current from the game view; a seat's
// badge reads it, so the board does not thread a prop through every
// panel.
export const voiceSeats: Writable<Record<string, VoiceConn>> = guardedWritable({}, "voiceSeats");

/** A seat's badge for the voice room, or null when it is not connected. */
export function voiceSeatBadge(
  conn: VoiceConn | undefined,
): { glyph: string; title: string } | null {
  if (conn === "live") return { glyph: "🎧", title: "in the table's voice room" };
  if (conn === "muted") return { glyph: "🔇", title: "in the table's voice room, muted" };
  return null;
}

// discordAppURL turns a room's https link into the one that opens the
// desktop app. Discord ignores it when it isn't installed, so callers
// fall back to the https link (voiceOpen).
export function discordAppURL(httpsURL: string): string | null {
  const m = /^https:\/\/discord\.com\/channels\/(\d+)\/(\d+)$/.exec(httpsURL);
  return m ? `discord://-/channels/${m[1]}/${m[2]}` : null;
}

// voiceOpen opens the room in the desktop app, and in the browser if the
// app doesn't take the link. The browser blurs the page when a protocol
// handler launches, which is the only signal of success we get.
export function voiceOpen(httpsURL: string): void {
  const app = discordAppURL(httpsURL);
  if (!app || typeof window === "undefined") {
    window.open(httpsURL, "_blank", "noopener,noreferrer");
    return;
  }
  let launched = false;
  const onBlur = () => (launched = true);
  window.addEventListener("blur", onBlur, { once: true });
  window.location.href = app;
  setTimeout(() => {
    window.removeEventListener("blur", onBlur);
    if (!launched) window.open(httpsURL, "_blank", "noopener,noreferrer");
  }, 1500);
}

// setVoiceRoom turns the table's voice room on or off (PUT
// /games/{id}/voice). Host or admin; the server refuses anyone else.
export async function setVoiceRoom(gameID: string, enabled: boolean): Promise<void> {
  await authFetch(`/games/${encodeURIComponent(gameID)}/voice`, {
    method: "PUT",
    body: JSON.stringify({ enabled }),
  });
}

// refreshVoiceRoom asks the server to check this seat's membership of
// the community server again, and to reopen a room removed for idleness
// (POST /games/{id}/voice/refresh). The answer arrives on the game view.
export async function refreshVoiceRoom(gameID: string): Promise<void> {
  await authFetch(`/games/${encodeURIComponent(gameID)}/voice/refresh`, {
    method: "POST",
    body: "{}",
  });
}

/** What the game bar shows for the voice room, or null for nothing. */
export interface VoiceChip {
  label: string;
  title: string;
  /** Open the room in Discord. Absent: the chip opens the voice panel. */
  href?: string;
}

// voiceChip is the bar's one-glance answer: a link straight into the
// room for a player who was let in, a nudge for one who still has a
// step to take, and nothing for a spectator or a table without voice.
export function voiceChip(v: VoiceView | undefined): VoiceChip | null {
  if (!v || !v.enabled) return null;
  const you = v.you ?? "spectator";
  if (you === "spectator") return null;
  if (v.status === "ready" && you === "granted" && v.url) {
    const count =
      v.granted !== undefined && v.granted > 0 ? ` · ${v.in_room ?? 0} of ${v.granted}` : "";
    return {
      label: `${v.in_room ? "Voice" : "Join voice"}${count}`,
      title: "open this table's voice room in Discord",
      href: v.url,
    };
  }
  if (you === "not_in_server") {
    return {
      label: "Voice: join the Discord",
      title: "join the CMD & CTRL Discord to use this table's voice room",
    };
  }
  if (you === "no_discord") {
    return {
      label: "Voice: sign in",
      title: "sign in with Discord to use this table's voice room",
    };
  }
  switch (v.status) {
    case "idle":
      return {
        label: "Voice: reopen",
        title: "the voice room was closed for inactivity — reopen it",
      };
    case "full":
      return { label: "Voice: full", title: "every voice room is in use right now" };
    case "unavailable":
      return {
        label: "Voice: unavailable",
        title: "Discord didn't answer; the game is unaffected",
      };
    default:
      return { label: "Voice: starting", title: "the voice room is being set up" };
  }
}

// voiceStatusLine is the panel's one-sentence account of the room.
export function voiceStatusLine(v: VoiceView): string {
  if (!v.enabled) return "No voice room for this table.";
  switch (v.status) {
    case "ready":
      return "This table has a private voice room in the CMD & CTRL Discord.";
    case "pending":
      return "The voice room opens once a player who signed in with Discord sits down.";
    case "idle":
      return "The voice room was closed after two quiet hours. Reopen it to talk again.";
    case "full":
      return "Every voice room is in use. One opens up when another table finishes.";
    case "unavailable":
      return "Discord didn't answer, so there's no voice room right now. The game is unaffected.";
    default:
      return "The voice room is closed.";
  }
}

// voiceYouLine tells this viewer what they can do, or null when there is
// nothing to say.
export function voiceYouLine(v: VoiceView): string | null {
  if (!v.enabled) return null;
  switch (v.you) {
    case "granted":
      return v.status === "ready" ? "You're in: open it in Discord." : null;
    case "not_in_server":
      return "You're not in the CMD & CTRL Discord yet. Join it, then check again.";
    case "no_discord":
      return "Guests can't join the voice room. Sign in with Discord to use it.";
    case "pending":
      return "Checking whether you're in the CMD & CTRL Discord…";
    default:
      return null;
  }
}
