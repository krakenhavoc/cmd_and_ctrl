// reveals turns GameView.reveals — the S22 broadcast reveal window
// (CR 701.20) — into the transient cues the board's attention strip
// shows.
//
// The server side is a WINDOW, not a stream: it carries up to four
// reveals from the current turn and reports them on every frame, so a
// client that dropped a frame still catches one. That is exactly the
// wrong shape to render directly. A window re-announces a reveal
// dozens of times as the turn goes on, and it would replay the whole
// turn's reveals at a client that has just reconnected.
//
// So the rules here are:
//
//   - a reveal is CUED the first time its `seq` is seen, and never
//     again. `seq` is an engine sequence number, monotonic and stable
//     across frames, which is the only thing on the entry that
//     identifies it — there are deliberately no card IDs to key on.
//   - a cue ages out on a timer, not on the window. It stops being
//     shown REVEAL_TTL_MS after it appeared, whether or not the
//     server has pushed it off the window yet, so the strip does not
//     hold a stale banner through a long turn.
//   - a client joining mid-game PRIMES: the window it finds on its
//     first frame is marked seen without being shown. Those reveals
//     already happened; replaying them as live announcements would be
//     a lie about when.
//
// Plain functions over plain data, so all of that is unit-testable
// without mounting a component.

import type { LogEvent, RevealView } from "./protocol";

// REVEAL_TTL_MS is how long one reveal stays on the attention strip.
//
// Long enough to read four or five card names off a banner without
// hurrying, short enough that it is gone before the next player's
// turn. The frame it came from outlives it — a reveal sits in the
// server's window for the rest of the turn — and that is the point of
// the split: the wire's job is that nobody MISSES the reveal, this
// timer's job is that nobody has to look at it forever.
export const REVEAL_TTL_MS = 9000;

// REVEAL_CUE_LIMIT caps how many reveal banners show at once. The
// strip is shared with the targeting banner, the bot feed and the
// toasts, so a resolution that reveals four times must not push the
// rest of it off the board.
export const REVEAL_CUE_LIMIT = 2;

// Random outcomes share the attention strip with reveals, but have a
// different payload: the server's `text` is the complete public line.
// Keep the cap and ageing policy aligned with reveals so a burst of
// rolls cannot crowd the board indefinitely.
export const RANDOM_CUE_LIMIT = 2;
export const RANDOM_CUE_TTL_MS = REVEAL_TTL_MS;

export interface RandomCue {
  log: LogEvent;
  shownAt: number;
}

export interface RandomState {
  seen: Set<number>;
  cues: RandomCue[];
}

export function emptyRandomState(): RandomState {
  return { seen: new Set(), cues: [] };
}

export function isRandomLog(log: LogEvent): boolean {
  return log.kind === "roll" || log.kind === "flip";
}

// Reconnecting clients prime the current log window. Those outcomes
// already happened before the client was watching, so they must not
// appear as fresh announcements.
export function primeRandomEvents(logs: readonly LogEvent[] | undefined): RandomState {
  const seen = new Set<number>();
  for (const log of logs ?? []) {
    if (isRandomLog(log)) seen.add(log.seq);
  }
  return { seen, cues: [] };
}

export function trackRandomEvents(
  prev: RandomState,
  logs: readonly LogEvent[] | undefined,
  now: number,
): RandomState {
  // The log window bounds memory. A shorter history after undo also
  // forgets rewound batches, allowing their replay to cue again.
  const current = logs === undefined ? undefined : new Set(logs.map((log) => log.seq));
  const seen = new Set([...prev.seen].filter((seq) => current === undefined || current.has(seq)));
  const newest = logs === undefined ? Infinity : Math.max(0, ...logs.map((log) => log.seq));
  const cues = prev.cues.filter((c) => c.log.seq <= newest && now - c.shownAt < RANDOM_CUE_TTL_MS);
  for (const log of logs ?? []) {
    if (!isRandomLog(log) || seen.has(log.seq)) continue;
    seen.add(log.seq);
    cues.push({ log, shownAt: now });
  }
  return { seen, cues: cues.slice(-RANDOM_CUE_LIMIT) };
}

export function dismissRandomEvent(prev: RandomState, seq: number): RandomState {
  return { seen: prev.seen, cues: prev.cues.filter((cue) => cue.log.seq !== seq) };
}

export interface RevealCue {
  reveal: RevealView;
  // shownAt is when this cue was first admitted, in ms.
  shownAt: number;
}

export interface RevealState {
  // seen is every reveal seq this client has already handled, cued or
  // primed. Monotonic, so it could be a high-water mark — but a set
  // costs nothing at four entries a turn and does not assume the
  // server never reorders.
  seen: Set<number>;
  cues: RevealCue[];
}

// emptyRevealState is the starting point for a client that has not
// received a frame yet.
export function emptyRevealState(): RevealState {
  return { seen: new Set(), cues: [] };
}

// primeRevealState marks a window as already handled without cueing
// any of it. Called once, with the first frame a client receives:
// whatever is in that window happened before this client was looking,
// and announcing it now would put a live banner on a reveal that may
// be a full turn old.
export function primeRevealState(reveals: RevealView[] | undefined): RevealState {
  const seen = new Set<number>();
  for (const r of reveals ?? []) seen.add(r.seq);
  return { seen, cues: [] };
}

// trackReveals folds one frame's window into the live cue list.
//
// `now` is passed rather than read from the clock so the ageing rule
// is testable. Returns a NEW state; the input is not mutated.
export function trackReveals(
  prev: RevealState,
  reveals: RevealView[] | undefined,
  now: number,
): RevealState {
  const seen = new Set(prev.seen);
  // Age out first, so a frame that admits nothing still expires what
  // is already up.
  const cues = prev.cues.filter((c) => now - c.shownAt < REVEAL_TTL_MS);

  for (const r of reveals ?? []) {
    if (seen.has(r.seq)) continue;
    seen.add(r.seq);
    cues.push({ reveal: r, shownAt: now });
  }
  // Newest wins the strip when a burst overflows it.
  return { seen, cues: cues.slice(-REVEAL_CUE_LIMIT) };
}

// dismissReveal drops one cue by seq. The reveal stays `seen`, so the
// next frame's window does not immediately put it back.
export function dismissReveal(prev: RevealState, seq: number): RevealState {
  return { seen: prev.seen, cues: prev.cues.filter((c) => c.reveal.seq !== seq) };
}

// hiddenRevealCount is how many cards a reveal showed that the frame
// had no room to name — Hermit Druid turning over a whole library.
// Zero for every ordinary reveal.
export function hiddenRevealCount(r: RevealView): number {
  const total = r.count ?? r.cards.length;
  return Math.max(0, total - r.cards.length);
}

// revealHeadline is the one line the banner leads with: who revealed,
// and what made them.
export function revealHeadline(r: RevealView, seatName: string): string {
  const who = seatName || "A player";
  if (r.source) return `${who} revealed · ${r.source}`;
  return `${who} revealed`;
}
