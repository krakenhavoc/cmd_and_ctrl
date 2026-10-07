// playmat.ts — ADR 0128. A signed-in person's playmat is one image the
// server stores and every player at the table sees behind THEIR
// battlefield, as with a paper mat.
//
// Pure logic only: which seats draw a mat under the per-device setting,
// and how a wire URL becomes an <img> source. No DOM, no network. The
// calls are in api.ts, the board layer in PlayerPanel.svelte and the
// Settings section in PlaymatSettings.svelte.

import { currentSession } from "./session";
import type { PlaymatsMode } from "./playmatMode";

export {
  DEFAULT_PLAYMATS_MODE,
  PLAYMATS_MODES,
  isPlaymatsMode,
  type PlaymatsMode,
} from "./playmatMode";

// The only shape the server mints. A wire value that is not exactly
// this is not loaded, whatever it says: the board must never be a way
// to make a browser contact a third-party host, which is the whole
// reason the server stores the bytes (ADR 0128 §3).
const PLAYMAT_PATH = /^\/playmats\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

/** isPlaymatPath reports whether url is a playmat route this server minted. */
export function isPlaymatPath(url: unknown): url is string {
  return typeof url === "string" && PLAYMAT_PATH.test(url);
}

/**
 * playmatSrc turns a wire playmat_url into the src an <img> loads, or
 * null for anything that is not a playmat path.
 *
 * The route is session-gated and an <img> cannot set an Authorization
 * header, so the token rides as ?token=, as avatarURL's does: the
 * session cookie alone is not reliable (a Secure-flag mismatch, a
 * cleared cookie with a live localStorage session).
 */
export function playmatSrc(url: string | undefined): string | null {
  if (!isPlaymatPath(url)) return null;
  const token = currentSession()?.token;
  return token ? `${url}?token=${encodeURIComponent(token)}` : url;
}

/**
 * playmatShownFor is the one rule for whether a seat's mat is drawn:
 * the setting, then whether the seat is the viewer's own. It returns
 * the wire URL, or null for none.
 */
export function playmatShownFor(
  mode: PlaymatsMode,
  seat: { playmat_url?: string },
  isSelf: boolean,
): string | null {
  if (mode === "off") return null;
  if (mode === "mine" && !isSelf) return null;
  return isPlaymatPath(seat.playmat_url) ? seat.playmat_url : null;
}

/** The server's cap for a pasted or uploaded image, for the form's hint. */
export const PLAYMAT_MAX_BYTES = 10 * 1024 * 1024;
