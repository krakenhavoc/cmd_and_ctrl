// seatNotch — the room a seated avatar keeps clear (ADR 0204).
//
// Every seat's identity sits at the middle of its hand's edge, in front
// of the fan (PlayerPanel.svelte). Two things must never end up under it:
//
//   - a permanent: the battlefield row next to the hand keeps a clear
//     band as deep as the identity reaches into it (§2, §3, call 3);
//   - one of the viewer's own hand cards: a card under the avatar dips
//     below it, so the fan's arc dips at its centre (§3, call 4).
//
// PlayerPanel measures the boxes and hands them here; these functions
// are the decisions, pinned by seatNotch.test.ts.

/** A vertical span in panel pixels. */
export interface VSpan {
  top: number;
  bottom: number;
}

/** A horizontal span in any one coordinate space. */
export interface HSpan {
  left: number;
  right: number;
}

/** The clear space kept between the identity and a row or a card. */
export const SEAT_CLEARANCE = 4;

export interface OverhangInput {
  /** The identity's drawn box: its top, and its lowest part (the life chip, badges). */
  identity: VSpan;
  /** The hand's row (the grid's bottom row, at the top edge on a top-row panel). */
  handRow: VSpan;
  /** The gap between grid rows. */
  rowGap: number;
  /** True when the hand is on the top edge (a top-row opponent). */
  handOnTop: boolean;
}

/**
 * seatOverhang is how far the identity reaches past its hand's row into
 * the battlefield next to it, plus the clearance: the depth of the band
 * that row keeps clear. Zero when the identity stays inside the hand's row.
 */
export function seatOverhang({ identity, handRow, rowGap, handOnTop }: OverhangInput): number {
  const reach = handOnTop
    ? identity.bottom + SEAT_CLEARANCE - (handRow.bottom + rowGap)
    : handRow.top - rowGap - (identity.top - SEAT_CLEARANCE);
  return Math.max(0, Math.round(reach));
}

/** Where the viewer's avatar sits over their fan, in the hand strip's x and y. */
export interface HandNotch {
  /** The avatar's horizontal extent (the portrait or its life chip, whichever is wider). */
  left: number;
  right: number;
  /** How far a card under it dips, in px. */
  depth: number;
}

export interface HandNotchInput {
  /** The portrait and the life chip, horizontally, in viewport px. */
  avatar: HSpan;
  /** The lowest drawn point of the identity, in viewport px. */
  avatarBottom: number;
  /** The top of the hand's strip, where a card at rest shows from, in viewport px. */
  handTop: number;
  /** The hand strip's left edge, in viewport px. */
  handLeft: number;
}

/**
 * handNotch converts the measured avatar into the hand strip's own x and
 * says how far a card under it must dip so its top shows below the
 * avatar's lowest point. Null when the avatar does not reach the hand.
 */
export function handNotch({
  avatar,
  avatarBottom,
  handTop,
  handLeft,
}: HandNotchInput): HandNotch | null {
  const depth = Math.round(avatarBottom + SEAT_CLEARANCE - handTop);
  if (depth <= 0 || avatar.right <= avatar.left) return null;
  return {
    left: Math.round(avatar.left - handLeft - SEAT_CLEARANCE),
    right: Math.round(avatar.right - handLeft + SEAT_CLEARANCE),
    depth,
  };
}

/** The gap the back row keeps from the viewer's avatar and its steppers. */
export const SEAT_BACK_GAP = 16;

/** The viewer's back row parted around their avatar: its grid columns and right padding, in px. */
export interface BackRowSplit {
  /** The lands' column, from the back row's left edge to the notch. */
  lands: number;
  /** The clear notch under the avatar and its steppers. */
  notch: number;
  /** The back row's right padding: the other permanents end at the dock. */
  right: number;
}

export interface BackRowSplitInput {
  panelW: number;
  /** Where the back row's columns start, in panel x (the panel's padding, and the coach card's). */
  start: number;
  /** The avatar with its satellites (mana pool, steppers), in panel x. */
  avatar: HSpan;
  /** The dock's left edge, in panel x. */
  dockLeft: number;
}

/**
 * backRowSplit (ADR 0204 §3, call 3) parts the viewer's back row around
 * their avatar where the two share height: on a wide board, where the
 * avatar straddles the line above the hand, and on a short one, where the
 * back row spills into the dock's row the avatar stands in. The lands go
 * left of the avatar, the other permanents right of it up to the dock,
 * and the notch between stays clear, so no permanent is under the avatar
 * and the creatures keep their height.
 */
export function backRowSplit({ panelW, start, avatar, dockLeft }: BackRowSplitInput): BackRowSplit {
  const notchLeft = avatar.left - SEAT_BACK_GAP;
  const notchRight = avatar.right + SEAT_BACK_GAP;
  return {
    lands: Math.max(0, Math.round(notchLeft - start)),
    notch: Math.max(0, Math.round(notchRight - Math.max(start, notchLeft))),
    right: Math.max(0, Math.round(panelW - dockLeft)),
  };
}

/**
 * slotDips is each hand card's dip: the notch's depth for a card whose
 * slot reaches under the avatar, else 0. `slots` are the slots' laid-out
 * spans in the strip's x (offsetLeft, before the fan's tilt), so a
 * tilted card's corner is allowed for by `tilt`.
 */
export function slotDips(slots: readonly HSpan[], notch: HandNotch | null, tilt = 6): number[] {
  return slots.map((s) =>
    notch && s.right + tilt > notch.left && s.left - tilt < notch.right ? notch.depth : 0,
  );
}
