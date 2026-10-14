// away.ts — the client's half of ADR 0177: who is away, and moving the
// table on without them.
//
// The server stamps each seat's `presence` (offline or hidden, and since
// when) and marks a seat `movable` once the table may move on without it.
// This file turns that into what the board shows (the chip), who is
// offered the choice (§3, mirrored; the server checks it again), the
// dock sheet that offers it, and the take-back banner. Everything here is
// pure, so a test can drive it with a frame and a clock.

import type { DockRequest } from "./dock";
import { L } from "./labels";
import type { GameView, PlayerView } from "./protocol";

/** ADR 0162's away time: a hidden tab shows "away" once it has lasted this long. */
export const HIDDEN_CHIP_AFTER_MS = 2 * 60 * 1000;

/** The away-time setting's default, in minutes (game.DefaultAwayAfter). */

export const DEFAULT_AWAY_AFTER = 3;

export type AwayAction = "stand_in" | "concede" | "wait";

/** Whole minutes since an RFC 3339 time, never negative; 0 for an unreadable one. */
export function minutesSince(since: string | undefined, now: number): number {
  const at = since ? Date.parse(since) : NaN;
  if (!Number.isFinite(at)) return 0;
  return Math.max(0, Math.floor((now - at) / 60000));
}

export interface PresenceChip {
  kind: "offline" | "away";
  text: string;
  title: string;
}

/**
 * The seat's presence chip, or null for none. Offline shows at once
 * ("offline 2m"); a hidden tab only once it has lasted two minutes
 * ("away 4m"), because a player glancing at Discord is still at the table.
 */
export function presenceChip(seat: PlayerView, now: number): PresenceChip | null {
  const p = seat.presence;
  if (!p || seat.is_bot || seat.eliminated) return null;
  const mins = minutesSince(p.since, now);
  if (p.state === "offline") {
    return {
      kind: "offline",
      text: `offline ${mins}m`,
      title: "Not connected to the table",
    };
  }
  if (p.state === "hidden") {
    const at = Date.parse(p.since);
    if (Number.isFinite(at) && now - at < HIDDEN_CHIP_AFTER_MS) return null;
    return {
      kind: "away",
      text: `away ${mins}m`,
      title: "Connected, but the table is not in front",
    };
  }
  return null;
}

/** The seats the table may move on without right now. */
export function movableSeats(view: GameView | null | undefined): PlayerView[] {
  return (view?.seats ?? []).filter((s) => s.presence?.movable === true && !s.stand_in);
}

/**
 * Whether the viewer may answer the offer for `away` (ADR 0177 §3): the
 * host while their own seat is present; otherwise any present seated
 * player still in the game; the server admin always. A mirror of the
 * server's check, so the sheet is only offered to someone it would accept.
 */
export function canDecideAway(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
  away: PlayerView,
  admin: boolean,
): boolean {
  if (admin) return true;
  if (!view || !viewerID || viewerID === away.id) return false;
  const me = view.seats.find((s) => s.id === viewerID);
  if (!me || me.is_bot || me.eliminated || me.presence?.state === "offline") return false;
  const host = view.seats.find((s) => s.is_host);
  if (host && host.id !== away.id && host.id !== viewerID && host.presence?.state !== "offline") {
    return false;
  }
  return true;
}

/** The first movable seat the viewer may decide for, or null. */
export function awayOfferFor(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
  admin: boolean,
): PlayerView | null {
  for (const s of movableSeats(view)) {
    if (canDecideAway(view, viewerID, s, admin)) return s;
  }
  return null;
}

function seatName(s: PlayerView): string {
  return s.display_name ?? s.name;
}

export interface AwayOfferOptions {
  seat: PlayerView;
  now: number;
  busy: boolean;
  error: string | null;
  onAnswer: (action: AwayAction) => void;
}

/** The dock sheet that offers the table a way on (ADR 0177 §2). */
export function awayOfferRequest(o: AwayOfferOptions): DockRequest {
  const name = seatName(o.seat);
  const mins = minutesSince(o.seat.presence?.since, o.now);
  return {
    // A choice, and the newest one: the table is waiting on the away
    // seat, so whatever else the decider has open (a mulligan sheet
    // waiting its turn) is waiting too, and must not hide the offer.
    rank: "choice",
    label: L.awayOffer,
    tag: "away",
    tone: "gold",
    question: `${name} has been offline for ${mins} minute${mins === 1 ? "" : "s"}.`,
    detail: "The table is waiting on them.",
    hint: o.error ?? "A bot plays their own deck and board, and they can take the seat back.",
    hintWarn: o.error !== null,
    row: [
      {
        id: "stand_in",
        label: L.awayStandIn(name),
        emphasis: true,
        disabled: o.busy,
        onPress: () => o.onAnswer("stand_in"),
      },
      {
        id: "concede",
        label: L.awayConcede(name),
        disabled: o.busy,
        onPress: () => o.onAnswer("concede"),
      },
      {
        id: "wait",
        label: L.awayWait,
        disabled: o.busy,
        onPress: () => o.onAnswer("wait"),
      },
    ],
    rowLayout: "wrap",
    primary: null,
    secondary: [],
  };
}

/** The viewer's own seat while a bot plays it, or null. */
export function standInSelf(
  view: GameView | null | undefined,
  viewerID: string | null | undefined,
): PlayerView | null {
  if (!view || !viewerID) return null;
  const me = view.seats.find((s) => s.id === viewerID);
  return me?.stand_in ? me : null;
}

/** "21:14", the stand-in's start in the viewer's own clock. */
export function standInSinceText(since: string | undefined): string {
  const at = since ? new Date(since) : null;
  if (!at || Number.isNaN(at.getTime())) return "a while ago";
  return at.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

/** The away-time setting in words, for the settings panel's hint. */
export function describeAwayAfter(n: number | undefined): string {
  const m = n ?? DEFAULT_AWAY_AFTER;
  if (m === 0) return "Never: the table always waits for a player who is gone.";
  return `When the table has waited on a disconnected player for ${m} minute${m === 1 ? "" : "s"}, the host can hand their seat to a bot, concede it for them, or keep waiting.`;
}
