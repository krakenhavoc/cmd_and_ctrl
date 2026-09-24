// Connection-state copy and input gating for the board (#519, ADR 0044).
//
// The whole of the connection UI used to be a nine-character pill in
// the command bar: a coloured dot and one word. Meanwhile the board
// stayed fully rendered and fully clickable, because the stale
// snapshot deliberately survives a reconnect — blanking the table on
// every network blip would be worse than showing a state that is a few
// seconds old. That trade is right. What was missing is the other half
// of it: the board has to ADMIT it is stale.
//
// So this module owns two things and no DOM:
//
//   1. Every sentence the client says about a dropped connection, in
//      one place, so it can be read end to end and tested as prose.
//   2. The predicates the UI gates on — is the board stale, are
//      actions live — so a route, a banner and a card back cannot
//      drift into three different answers.
//
// The `ConnectionStatus` import is type-only and therefore erased at
// build time. That matters: ws.ts imports OFFLINE_ERROR_CODE and
// offlineSendMessage from HERE, so turning this into a value import
// would close a genuine runtime cycle. It stays `import type`.
import { SIGNED_IN_RETURN_LINK } from "./joinRecovery";
import type { ConnectionStatus } from "./ws";

// OFFLINE_ERROR_CODE is the `code` on the lastError entry raised when
// something is sent at a socket that is not open. No server sends it —
// it is minted client-side — but it travels the same store and renders
// in the same toast as a server rejection, which is the point: the
// player should not have to know which half of the system said no.
export const OFFLINE_ERROR_CODE = "not_connected";

// STALE_BOARD_SENTENCE is the one claim every disconnected surface
// makes, and the reason this module exists. It names what the player
// is looking at rather than asking them to infer it from a dot.
export const STALE_BOARD_SENTENCE =
  "The board below is the last state the server sent, not live play.";

// FREEZE_WORDS is the vocabulary that belongs to #266 — the real
// board-freeze bug, where the client throws mid-render and the table
// stops updating with the socket perfectly healthy.
//
// Nothing in this module may use those words. The two failures look
// identical to a player and have nothing in common underneath, so if
// the disconnect banner says "frozen" then every genuine freeze gets
// reported as a network problem and stops being findable. Enforced by
// connectionBanner.test.ts rather than by reviewer memory.
export const FREEZE_WORDS = [
  "freeze",
  "froze",
  "frozen",
  "stuck",
  "hung",
  "crash",
  "not responding",
] as const;

// usesFreezeVocabulary reports whether `text` reaches for #266's
// words. Case-insensitive; substring rather than word-boundary
// matching, so "crashed" and "freezes" are caught too.
export function usesFreezeVocabulary(text: string): boolean {
  const lower = text.toLowerCase();
  return FREEZE_WORDS.some((word) => lower.includes(word));
}

// offlineSendMessage is the toast text for a frame that could not be
// sent because the socket is down. `what` names the frame in the
// player's terms — `action "pass_priority"`, `chat message`, `ping`.
//
// It says three things in order, all of which were previously
// unsaid: the connection is the problem, the specific thing you just
// did did not happen, and what you are still looking at.
export function offlineSendMessage(what: string): string {
  return `Not connected — your ${what} was not sent. ${STALE_BOARD_SENTENCE}`;
}

// A banner's tone. "retrying" is the client still working the backoff
// ladder; "lost" is a connection that will not come back on its own
// but might, later, with no action from the player; "ended" (#1475)
// is the one banner that will NEVER clear itself — the session is
// gone and only signing in again fixes it.
export type BannerTone = "retrying" | "lost" | "ended";

export interface ConnectionBannerState {
  tone: BannerTone;
  /** Short enough to read at a glance from across the table. */
  headline: string;
  /** What is on screen and what is not happening. */
  detail: string;
  /** Label for the manual retry control. Empty when `link` is set instead. */
  retryLabel: string;
  /**
   * Where to send the player instead of retrying (#1475's "ended"
   * tone only). A dead session cannot be fixed by dialling harder, so
   * this replaces the retry button with a real navigation.
   */
  link?: { href: string; label: string };
}

// boardIsStale reports whether what is rendered may be out of date —
// the condition the banner and any "stale" board treatment gate on.
//
// "connecting" is deliberately NOT stale: it is the first dial of a
// fresh mount, there is no snapshot yet, and the route already renders
// its own "waiting for snapshot…" state for that. Marking it stale
// would put a scary banner on every normal page load.
export function boardIsStale(status: ConnectionStatus): boolean {
  return status === "reconnecting" || status === "disconnected" || status === "session_ended";
}

// actionsDisabled reports whether action affordances should be shown
// as unavailable. Broader than boardIsStale on purpose: during the
// first "connecting" dial there is genuinely no socket to send on
// either, so a control that looks live would lie.
//
// This is the predicate the board's own input gating should read
// (#519's remaining task — see the PR body); sendAction refuses on its
// own regardless, so a click that slips through is surfaced rather
// than swallowed.
export function actionsDisabled(status: ConnectionStatus): boolean {
  return status !== "connected";
}

// attemptPhrase renders the reconnect attempt count #518 exposes.
// Defensive about the input because it is read straight off a store:
// a zero, a negative or a NaN means "we have not counted a retry yet"
// and simply drops the parenthetical rather than rendering "attempt
// NaN" at a player mid-game.
function attemptPhrase(attempt: number): string {
  if (!Number.isFinite(attempt) || attempt < 1) return "Retrying automatically.";
  return `Retrying automatically (attempt ${Math.floor(attempt)}).`;
}

// connectionBanner returns what the banner should say, or null when
// there is nothing to say. Pure — the component renders this and holds
// no copy of its own.
//
// `signedIn` only matters for the "session_ended" status (#1475): by
// the time that status lands, the dead session itself has usually
// already been cleared (authFetch's own 401 handling), so the caller
// passes what it last knew about who was signed in rather than this
// module reading a session store that may already be empty.
export function connectionBanner(
  status: ConnectionStatus,
  attempt: number,
  signedIn = false,
): ConnectionBannerState | null {
  if (!boardIsStale(status)) return null;
  if (status === "reconnecting") {
    return {
      tone: "retrying",
      headline: "Connection lost — reconnecting",
      detail: `${STALE_BOARD_SENTENCE} ${attemptPhrase(attempt)}`,
      retryLabel: "Try now",
    };
  }
  if (status === "session_ended") {
    return {
      tone: "ended",
      headline: "Your session ended — sign in again",
      detail: signedIn
        ? `${STALE_BOARD_SENTENCE} Head to My games to get back to an open table.`
        : `${STALE_BOARD_SENTENCE} Nothing you do will reach the game until you sign in again.`,
      retryLabel: "",
      link: signedIn ? SIGNED_IN_RETURN_LINK : { href: "#/login", label: "Sign in" },
    };
  }
  return {
    tone: "lost",
    headline: "Disconnected from the table",
    detail: `${STALE_BOARD_SENTENCE} Nothing you do will reach the game until the connection is back.`,
    retryLabel: "Reconnect",
  };
}

// connectionAnnouncement is the line read into the aria-live region on
// a status change. Separate from the banner copy because it covers all
// five states, not just the three that render a banner: a screen-reader
// user needs to be told the table came BACK at least as much as they
// need to be told it went away, and the banner says that by vanishing
// silently.
export function connectionAnnouncement(status: ConnectionStatus, attempt: number): string {
  switch (status) {
    case "connected":
      return "Reconnected. The board is live again.";
    case "connecting":
      return "Connecting to the table.";
    case "reconnecting":
      return `Connection lost. ${attemptPhrase(attempt)} The board is not live.`;
    case "disconnected":
      return "Disconnected from the table. The board is not live.";
    case "session_ended":
      return "Your session ended. Sign in again to get back to the table.";
  }
}
