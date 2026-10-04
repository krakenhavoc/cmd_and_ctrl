// boardExpand — a seat's board drawn larger over the table (ADR 0120).
//
// Hovering a seat's avatar opens a PEEK of that seat's board in an
// overlay beside it; clicking the avatar (when it is not picking the
// player as a target or an attack) PINS it, and it stays until it is
// unpinned. This module is the whole set of transitions, as one pure
// reducer, so every rule in ADR 0120 §1 is a unit test rather than a
// reading of Board.svelte. Board owns the state, feeds it events, and
// runs the one timer the state asks for.
//
// The timer is part of the state (`timer`), with a sequence number. A
// timer that fires reports its `seq`; a fire whose seq is no longer the
// state's is stale and changes nothing. That keeps "the pointer came
// back inside the grace" a property of the reducer, not of whether
// Board remembered to clear a setTimeout.
//
// Placement (ADR 0120 §2) is here too, as a pure function over two
// rectangles, so "never over the avatar" is tested without a layout
// engine.

import { getContext, setContext } from "svelte";
import type { Readable } from "svelte/store";
import { guardedWritable } from "./guardedStore";

/** The open overlay: whose board, and whether it is pinned. */
export interface Expanded {
  seatID: string;
  pinned: boolean;
}

/** The one timer the state wants running, if any. */
export interface ExpandTimer {
  kind: "open" | "close";
  ms: number;
  /** Bumped on every (re)schedule; a fire with an older seq is stale. */
  seq: number;
}

export interface ExpandState {
  expanded: Expanded | null;
  /** The seat whose avatar the pointer rests on, or null. */
  overAvatar: string | null;
  /** Whether the pointer is inside the overlay. */
  overOverlay: boolean;
  /** A peek waiting on the open timer, for this seat. */
  pendingSeat: string | null;
  /**
   * An avatar whose hover was used up — by a press, or by an unpin or
   * Escape while the pointer rested on it. It opens no peek until the
   * pointer has left it (ADR 0120 §1, "A press consumes the hover").
   */
  consumed: string | null;
  timer: ExpandTimer | null;
  /** The last seq handed out, so a new timer always gets a fresh one. */
  seq: number;
}

/** ADR 0120 §1 and Calls made here, item 5. */
export const PEEK_OPEN_MIN_MS = 400;
export const PEEK_CLOSE_GRACE_MS = 250;

/** The open delay: never under 400 ms, and slower if the zoom is. */
export function peekOpenDelay(hoverDelayMs: number): number {
  return Math.max(PEEK_OPEN_MIN_MS, Number.isFinite(hoverDelayMs) ? hoverDelayMs : 0);
}

export type ExpandEvent =
  /** The pointer came onto a seat's avatar (hover-capable, no button held). */
  | { type: "avatar-enter"; seatID: string; openDelayMs: number }
  /** The pointer left a seat's avatar. */
  | { type: "avatar-leave"; seatID: string }
  /** A pointerdown on a seat's avatar. */
  | { type: "avatar-press"; seatID: string }
  /**
   * A click on a seat's avatar that neither targeted nor attacked the
   * player (those intercepts run first, in PlayerIdentity).
   */
  | { type: "avatar-click"; seatID: string }
  | { type: "overlay-enter" }
  | { type: "overlay-leave" }
  /** The overlay's pin button. */
  | { type: "pin-toggle" }
  /** An expand button (a full panel's, or a summary's ⤢ or pip). */
  | { type: "expand"; seatID: string }
  /** Escape, once nothing else owns it. */
  | { type: "escape" }
  /** A timer fired. */
  | { type: "elapsed"; seq: number };

export function initialExpandState(): ExpandState {
  return {
    expanded: null,
    overAvatar: null,
    overOverlay: false,
    pendingSeat: null,
    consumed: null,
    timer: null,
    seq: 0,
  };
}

function schedule(s: ExpandState, kind: ExpandTimer["kind"], ms: number): ExpandState {
  const seq = s.seq + 1;
  return { ...s, seq, timer: { kind, ms, seq } };
}

function cancel(s: ExpandState, kind?: ExpandTimer["kind"]): ExpandState {
  if (!s.timer || (kind && s.timer.kind !== kind)) return s;
  return { ...s, timer: null };
}

// A peek stays while the pointer is on its avatar or in the overlay.
function peekHeld(s: ExpandState): boolean {
  return s.overOverlay || (s.expanded !== null && s.overAvatar === s.expanded.seatID);
}

// After the pointer moved: a peek nobody holds starts its grace; one
// that is held again cancels it.
function settlePeek(s: ExpandState): ExpandState {
  const e = s.expanded;
  if (!e || e.pinned) return cancel(s, "close");
  if (peekHeld(s)) return cancel(s, "close");
  if (s.timer?.kind === "close") return s;
  return schedule(s, "close", PEEK_CLOSE_GRACE_MS);
}

// Closing: the overlay goes, and an avatar the pointer rests on is used
// up, so it does not peek straight back open under the pointer.
function close(s: ExpandState): ExpandState {
  return {
    ...s,
    expanded: null,
    overOverlay: false,
    pendingSeat: null,
    timer: null,
    consumed: s.overAvatar ?? s.consumed,
  };
}

/** reduceExpand is every transition in ADR 0120 §1. */
export function reduceExpand(s: ExpandState, ev: ExpandEvent): ExpandState {
  switch (ev.type) {
    case "avatar-enter": {
      let next: ExpandState = { ...s, overAvatar: ev.seatID };
      // Its own peek is held again: cancel any grace.
      if (next.expanded?.seatID === ev.seatID) return settlePeek({ ...next, pendingSeat: null });
      // There is never more than one overlay: no peek while pinned.
      if (next.expanded?.pinned) return next;
      if (next.consumed === ev.seatID) return next;
      next = { ...next, pendingSeat: ev.seatID };
      return schedule(next, "open", ev.openDelayMs);
    }
    case "avatar-leave": {
      if (s.overAvatar !== ev.seatID && s.pendingSeat !== ev.seatID && s.consumed !== ev.seatID) {
        return s;
      }
      let next: ExpandState = {
        ...s,
        overAvatar: s.overAvatar === ev.seatID ? null : s.overAvatar,
        consumed: s.consumed === ev.seatID ? null : s.consumed,
      };
      if (next.pendingSeat === ev.seatID) {
        next = cancel({ ...next, pendingSeat: null }, "open");
      }
      return settlePeek(next);
    }
    case "avatar-press": {
      let next: ExpandState = { ...s, consumed: ev.seatID };
      if (next.pendingSeat !== null) next = cancel({ ...next, pendingSeat: null }, "open");
      return next;
    }
    case "avatar-click": {
      const e = s.expanded;
      if (e?.seatID === ev.seatID && e.pinned) {
        return { ...close(s), consumed: ev.seatID };
      }
      return {
        ...cancel(s),
        expanded: { seatID: ev.seatID, pinned: true },
        pendingSeat: null,
        consumed: ev.seatID,
      };
    }
    case "overlay-enter":
      return settlePeek({ ...s, overOverlay: true });
    case "overlay-leave":
      return settlePeek({ ...s, overOverlay: false });
    case "pin-toggle": {
      const e = s.expanded;
      if (!e) return s;
      if (e.pinned) return close(s);
      return { ...cancel(s), expanded: { ...e, pinned: true }, pendingSeat: null };
    }
    case "expand":
      return {
        ...cancel(s),
        expanded: { seatID: ev.seatID, pinned: true },
        pendingSeat: null,
      };
    case "escape":
      return s.expanded ? close(s) : s;
    case "elapsed": {
      const t = s.timer;
      if (!t || t.seq !== ev.seq) return s;
      const next: ExpandState = { ...s, timer: null };
      if (t.kind === "open") {
        const seat = next.pendingSeat;
        if (!seat || next.expanded?.pinned) return { ...next, pendingSeat: null };
        // The pointer is still on that avatar, so the new peek is held.
        return { ...next, pendingSeat: null, expanded: { seatID: seat, pinned: false } };
      }
      // The close grace ran out with nobody holding the peek.
      if (next.expanded && !next.expanded.pinned && !peekHeld(next)) return close(next);
      return next;
    }
  }
}

/**
 * liveExpanded is the open overlay, or null when its seat is no longer
 * at the table. A derivation, not an effect (ADR 0120 §1, "Gone seat,
 * gone overlay"), exactly as ADR 0077's pin was dropped.
 */
export function liveExpanded(e: Expanded | null, seatIDs: Iterable<string>): Expanded | null {
  if (!e) return null;
  for (const id of seatIDs) if (id === e.seatID) return e;
  return null;
}

// ---------------------------------------------------------------- //
// Placement (ADR 0120 §2)

/** A box in board-relative pixels. */
export interface Box {
  left: number;
  top: number;
  width: number;
  height: number;
}

export interface PlacementOptions {
  /** The board's width in pixels. */
  boardWidth: number;
  /** ADR 0119's pile clearance on the left (`--stack-pile-clear`). */
  leftClear?: number;
  /** The gap between the avatar and the overlay. */
  gap?: number;
  /** The widest the overlay gets. */
  maxWidth?: number;
  /** The inset from the board's edges (its padding). */
  edge?: number;
}

export const OVERLAY_GAP = 8;
export const OVERLAY_MAX_WIDTH = 1200;

/** The horizontal span the overlay takes: board-relative left and width. */
export interface Span {
  left: number;
  width: number;
  side: "left" | "right";
}

/**
 * placeOverlay takes the wider of the two spans beside the avatar, each
 * capped and hugging the avatar, so the overlay never covers it and the
 * pointer only has the gap to cross. `avatar` is board-relative.
 */
export function placeOverlay(avatar: Box, opts: PlacementOptions): Span {
  const gap = opts.gap ?? OVERLAY_GAP;
  const cap = opts.maxWidth ?? OVERLAY_MAX_WIDTH;
  const edge = opts.edge ?? 0;
  const leftEdge = edge + Math.max(0, opts.leftClear ?? 0);
  const rightEdge = opts.boardWidth - edge;

  const leftRoom = Math.max(0, avatar.left - gap - leftEdge);
  const rightRoom = Math.max(0, rightEdge - (avatar.left + avatar.width + gap));
  const leftW = Math.min(cap, leftRoom);
  const rightW = Math.min(cap, rightRoom);
  if (leftW >= rightW) {
    return { left: avatar.left - gap - leftW, width: leftW, side: "left" };
  }
  return { left: avatar.left + avatar.width + gap, width: rightW, side: "right" };
}

/** parsePx reads a CSS length in px ("212px", "0", ""), or 0. */
export function parsePx(v: string | null | undefined): number {
  const n = Number.parseFloat((v ?? "").trim());
  return Number.isFinite(n) ? n : 0;
}

// ---------------------------------------------------------------- //
// Layout tick

const layout = guardedWritable(0, "boardExpandLayout");

/**
 * boardExpandLayout ticks whenever the overlay opens, closes, changes
 * seat or resizes. CombatArrows and the fan lane read it to re-measure
 * (ADR 0120 §3): the overlay's copy of a card is the anchor while it is
 * open, and moving it does not change the snapshot.
 */
export const boardExpandLayout: Readable<number> = { subscribe: layout.subscribe };

/** bumpBoardExpandLayout is Board's half of boardExpandLayout. */
export function bumpBoardExpandLayout(): void {
  layout.update((n) => n + 1);
}

// ---------------------------------------------------------------- //
// The avatar's half (PlayerIdentity.svelte)

/**
 * What an avatar reports to the board that owns the overlay. Board sets
 * it as context; an avatar rendered with no Board above it (a unit test,
 * a lobby preview) has none and simply does not expand.
 */
export interface AvatarExpandHandlers {
  enter(seatID: string, ev: PointerEvent): void;
  leave(seatID: string): void;
  press(seatID: string): void;
  /** A click the target and attack intercepts did not take. */
  click(seatID: string): void;
}

const AVATAR_KEY = Symbol("avatar-expand");

/** setAvatarExpand is Board's: call it during component init. */
export function setAvatarExpand(h: AvatarExpandHandlers): void {
  setContext(AVATAR_KEY, h);
}

/** avatarExpand is the avatar's: call it during component init. */
export function avatarExpand(): AvatarExpandHandlers | undefined {
  return getContext<AvatarExpandHandlers | undefined>(AVATAR_KEY);
}
