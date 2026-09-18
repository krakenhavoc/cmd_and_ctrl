// lifePopup turns a seat's `life_history` — the server's rolling
// per-player life-change log — into the floating ±N cue over that
// seat's identity disc (#703).
//
// The wire shape is a BOUNDED LOG, not a stream of notifications: it
// carries the newest MaxLifeHistoryEntries (50) changes and is
// re-sent whole on every frame. Two things follow, and the old
// PlayerIdentity code got both wrong by diffing the log's LENGTH
// against a baseline:
//
//   - a frame can carry SEVERAL new changes, and a length diff only
//     ever surfaces the last one. Double strike to a player lands both
//     combat damage steps' entries in one frame, so a 3/3 double
//     striker showed "-3" and took 6. Lifelink plus damage, and drain
//     triggers, do the same thing.
//   - once the log hits its cap the length stops growing, so
//     `count <= baselineCount` was true forever after and NO popup
//     fired again for the rest of the game.
//
// So a popup is keyed on `seq` — the per-player counter the server
// stamps on each entry, which keeps climbing after the log stops
// growing (the reveals.ts / combatBeats.ts watermark pattern). The
// rules:
//
//   - the first frame a client sees PRIMES: everything already on the
//     wire is marked seen and nothing pops. Those changes already
//     happened; floating them now would be a lie about when. A
//     remount (reconnect, board re-render) primes again for the same
//     reason.
//   - a frame whose highest seq is BELOW the watermark is a rewind
//     (undo, or a replay scrubbed backwards): the watermark drops to
//     it and nothing pops, so the replayed changes pop when they come
//     back.
//   - every unseen non-zero delta in the frame is shown, oldest
//     first — not summed. The seat's life NUMBER is already on screen
//     directly under the popup, so the resulting total is never the
//     missing information; the individual beats are. Summing would
//     throw away the only thing the popup adds.
//   - beyond LIFE_POPUP_MAX_LINES deltas a stack stops being
//     readable, so it collapses to the net total plus the count. The
//     chip still carries the resulting total either way.
//   - when the log was trimmed past the watermark, some changes are
//     genuinely gone from the wire. The popup says so (`gap`) rather
//     than presenting what survived as if it were everything.
//
// Plain functions over plain data, so all of it is unit-testable
// without mounting a component (#689). PlayerIdentity.svelte is the
// thin shell that holds the tracker and renders the lines.

import type { LifeChangeView } from "./protocol";

// LIFE_POPUP_MAX_LINES is how many deltas are shown individually
// before the popup collapses to a net total. Four covers every case
// the rules actually produce in one frame — two combat damage steps,
// a damage-plus-lifelink pair, a drain — with room to spare; past
// that it is a wall of numbers over a 88px disc.
export const LIFE_POPUP_MAX_LINES = 4;

// LIFE_POPUP_HOLD_MS is how long a single-delta popup stays up, and
// LIFE_POPUP_HOLD_PER_EXTRA_MS is what each further line adds: a
// stack of four numbers takes longer to read than one. Capped at
// LIFE_POPUP_HOLD_MAX_MS so the cue never outlives the moment.
//
// Not speed-scaled: this is a number to read, not motion, so the
// animation-speed setting must not make it unreadable (the
// combatBeats.ts BEAT_CUE_HOLD_MS rule).
export const LIFE_POPUP_HOLD_MS = 1100;
export const LIFE_POPUP_HOLD_PER_EXTRA_MS = 450;
export const LIFE_POPUP_HOLD_MAX_MS = 3000;

// LifeTracker is one seat's watermark over the life log.
export interface LifeTracker {
  primed: boolean;
  // Highest entry seq already shown. Entries are appended in seq
  // order, so a watermark is the same as a seen set, and a rewind is
  // just lowering it.
  maxSeq: number;
}

export function emptyLifeTracker(): LifeTracker {
  return { primed: false, maxSeq: 0 };
}

export interface LifePopup {
  // Stable key for a keyed block: the seq of the newest delta in it.
  // Restarts the in-transition when a genuinely new change lands, and
  // does not when the same popup is re-derived from a later frame.
  key: number;
  // Every unseen non-zero delta, oldest first.
  deltas: number[];
  // What the life total moved by across all of them.
  total: number;
  // The life total after the last of them, as the server reported it.
  newTotal: number;
  // True when the log was trimmed past the watermark, so changes
  // between what was last shown and the oldest surviving entry are
  // gone from the wire and cannot be shown.
  gap: boolean;
}

export interface LifeFrame {
  tracker: LifeTracker;
  // The popup this frame raises, or null when it raises none (it
  // primed, it rewound, or it carried no new change). Null never
  // means "take the current popup down" — a popup ages out on its own
  // hold timer.
  popup: LifePopup | null;
  primed: boolean;
  rewound: boolean;
}

function highestSeq(history: readonly LifeChangeView[] | null | undefined): number {
  let top = 0;
  for (const e of history ?? []) if (e.seq > top) top = e.seq;
  return top;
}

// trackLife folds one frame's life log into the tracker. Returns a new
// tracker; the input is not mutated.
export function trackLife(
  prev: LifeTracker,
  history: readonly LifeChangeView[] | null | undefined,
): LifeFrame {
  const entries = history ?? [];
  const top = highestSeq(entries);

  if (!prev.primed) {
    return { tracker: { primed: true, maxSeq: top }, popup: null, primed: true, rewound: false };
  }
  if (top < prev.maxSeq) {
    // An undo rewound the log, or a replay was scrubbed backwards.
    return { tracker: { primed: true, maxSeq: top }, popup: null, primed: false, rewound: true };
  }

  const fresh = entries.filter((e) => e.seq > prev.maxSeq && e.delta !== 0);
  if (fresh.length === 0) {
    return {
      tracker: { primed: true, maxSeq: Math.max(top, prev.maxSeq) },
      popup: null,
      primed: false,
      rewound: false,
    };
  }

  // Oldest first. The server appends in seq order, but sort rather
  // than trust it: the cost is nothing and a reordered frame would
  // otherwise show the deltas backwards.
  const ordered = [...fresh].sort((a, b) => a.seq - b.seq);
  const newest = ordered[ordered.length - 1];
  // A gap means entries were trimmed off the front between the
  // watermark and what this frame still carries. `prev.maxSeq + 1` is
  // the seq the next change we should have seen would have had.
  const gap = ordered[0].seq > prev.maxSeq + 1;

  return {
    tracker: { primed: true, maxSeq: top },
    popup: {
      key: newest.seq,
      deltas: ordered.map((e) => e.delta),
      total: ordered.reduce((sum, e) => sum + e.delta, 0),
      newTotal: newest.new_total,
      gap,
    },
    primed: false,
    rewound: false,
  };
}

export type LifeTone = "loss" | "gain";

export interface LifePopupLine {
  // "-3" / "+2", already signed for display.
  text: string;
  tone: LifeTone;
}

export interface LifePopupView {
  // What to render, top to bottom: the deltas oldest-first, or one
  // collapsed net total.
  lines: LifePopupLine[];
  // True when `lines` is the collapsed net total rather than the
  // individual deltas; `count` is how many changes it stands for.
  collapsed: boolean;
  count: number;
  // Changes are missing from the wire (see LifePopup.gap).
  gap: boolean;
  // The whole popup as one sentence for the aria-live region: a
  // screen reader gets one announcement, not one per line.
  label: string;
  // Which cue to play. Any loss in the group plays "damage" even
  // alongside a gain — taking damage is the part a player must not
  // miss.
  sound: "damage" | "heal";
  holdMs: number;
}

function signed(delta: number): string {
  return delta > 0 ? `+${delta}` : `${delta}`;
}

function toneOf(delta: number): LifeTone {
  return delta < 0 ? "loss" : "gain";
}

function phrase(delta: number): string {
  return delta < 0 ? `lost ${-delta}` : `gained ${delta}`;
}

// lifePopupLabel is the one-sentence announcement for the popup.
// Words, not glyphs: "-3" is read inconsistently, "lost 3 life" is
// not.
export function lifePopupLabel(popup: LifePopup): string {
  const parts: string[] = [];
  if (popup.deltas.length === 1) {
    parts.push(`${phrase(popup.deltas[0])} life`);
  } else {
    const each = popup.deltas.map(phrase).join(", ");
    const net = popup.total === 0 ? "no net change" : `${phrase(popup.total)} life in all`;
    parts.push(`${each} — ${net}`);
  }
  parts.push(`now ${popup.newTotal}`);
  if (popup.gap) parts.push("earlier changes not shown");
  return parts.join(", ");
}

// lifePopupView is everything the component needs to render one
// popup. Split from trackLife so the presentation choice — stack the
// deltas, collapse a long run — is testable on its own.
export function lifePopupView(popup: LifePopup, maxLines = LIFE_POPUP_MAX_LINES): LifePopupView {
  const count = popup.deltas.length;
  const collapsed = count > Math.max(1, maxLines);
  const lines: LifePopupLine[] = collapsed
    ? [{ text: signed(popup.total), tone: toneOf(popup.total) }]
    : popup.deltas.map((d) => ({ text: signed(d), tone: toneOf(d) }));
  return {
    lines,
    collapsed,
    count,
    gap: popup.gap,
    label: lifePopupLabel(popup),
    sound: popup.deltas.some((d) => d < 0) ? "damage" : "heal",
    holdMs: Math.min(
      LIFE_POPUP_HOLD_MAX_MS,
      LIFE_POPUP_HOLD_MS + (lines.length - 1) * LIFE_POPUP_HOLD_PER_EXTRA_MS,
    ),
  };
}
