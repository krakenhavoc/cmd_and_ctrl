// stackLinger — the linger on resolution (ADR 0119 §3).
//
// When an item leaves the stack the board's next frame no longer
// draws it, so a spell that resolved, was countered or fizzled used to
// vanish with nothing to say which. The linger keeps a copy of the
// departed card where it was drawn for about 1.5 s, badged with what
// happened to it, then sends it toward where it went.
//
// DISPLAY ONLY. Nothing waits for it: the board already shows the
// frame's end state and input stays live under it, as for ADR 0053's
// combat beats.
//
// THE SIGNAL is the public log, never a guess. An item lingers only
// when a `resolve`, `fizzle` or `counter` entry newer than the last one
// this client handled names it (`stack_item_id`, ADR 0119 PR 4; a
// spell's `card_id` or a counter's `target` from a server that predates
// it, which are the same IDs for a spell). An item that leaves with no
// such entry — an undo, a spell returned to its owner's hand, a
// reconnect — fades out in 200 ms with no badge, because a badge would
// be a claim the client cannot back.
//
// PRIMING. The first frame after mounting, a reconnect or a replay
// toggle primes the tracker and lingers nothing, exactly as the combat
// beats do (`track` in combatBeats.ts is reused for the log watermark,
// so "which entries are new" and "an undo rewinds" mean the same thing
// in both places).
//
// Plain functions over plain data, plus one small timer class, so all
// of it is unit-testable without mounting a component (#689).
// StackLinger.svelte measures, draws and flies.

import type { GameView, LogEvent, StackItemView } from "./protocol";
import type { StackLaneModel } from "./stackLane";
import { cardImageURL } from "./cardImage";
import { cardSelector } from "./boardAnchor";
import { emptyBeatTracker, track, type BeatTracker } from "./combatBeats";

export type LingerOutcome = "resolved" | "countered" | "fizzled";

/** How long a lone departed card lingers. Reading time, so never speed-scaled. */
export const LINGER_MS = 1500;
/** The shortest linger a card in a burst gets. */
export const LINGER_BURST_MIN_MS = 400;
/** At most this many lingers are queued, the one on screen included. */
export const LINGER_QUEUE_MAX = 3;
/** An item that left with no log entry fades out in this long, with no badge. */
export const LINGER_FADE_MS = 200;
/** The flight to where the card went, before animations.speed. */
export const LINGER_FLIGHT_MS = 400;

/** A box in board pixels. */
export interface Box {
  left: number;
  top: number;
  width: number;
  height: number;
}

/** What the board drew for one stack item, kept so it can linger after it has gone. */
export interface LingerItem {
  id: string;
  kind: StackItemView["kind"];
  /** The verbatim title, as the stack drew it. */
  name: string;
  /** The `normal` image the item was drawn as, a card back, or null for no art. */
  image: string | null;
  /** The `small` image, for a row-shaped copy. */
  thumb: string | null;
  casterColor: string;
  /** CR 707.10: a copy of a spell ceases to exist when it leaves the stack. */
  isCopy: boolean;
}

/** One item that left the stack between two frames. */
export interface Departure {
  item: LingerItem;
  /** What the log says happened to it; null when no entry names it. */
  outcome: LingerOutcome | null;
  /** For a counter, the name of what countered it ("Counterspell"). */
  by: string | null;
  /** The naming entry's seq, or -1 with no entry. */
  seq: number;
  /** Where the item was drawn in the previous frame, or null if it was not drawn. */
  rect: Box | null;
}

export interface LingerTracker {
  log: BeatTracker;
  /** The previous frame's stack items, by ID. */
  items: Map<string, LingerItem>;
}

export function emptyLingerTracker(): LingerTracker {
  return { log: emptyBeatTracker(), items: new Map() };
}

export interface LingerFrame {
  tracker: LingerTracker;
  /** Items to linger with a badge, in resolution (log) order. */
  lingers: Departure[];
  /** Items that left with no entry naming them: a 200 ms fade, no badge. */
  fades: Departure[];
  /** True when this frame primed: nothing lingers or fades. */
  primed: boolean;
}

const OUTCOME_OF: Partial<Record<LogEvent["kind"], LingerOutcome>> = {
  resolve: "resolved",
  counter: "countered",
  fizzle: "fizzled",
};

/**
 * The stack item a resolve, fizzle or counter entry is about, or null
 * for any other entry. `stack_item_id` when the server sends it; else a
 * spell's `card_id` (resolve, fizzle) or a counter's `target`, which
 * for a spell are its item's ID. An ability's `card_id` is its source
 * permanent, which is never a stack item's ID, so the fallback cannot
 * mis-match one.
 */
export function entryItemID(e: LogEvent): string | null {
  if (!OUTCOME_OF[e.kind]) return null;
  if (e.stack_item_id) return e.stack_item_id;
  const id = e.kind === "counter" ? e.target : e.card_id;
  return id || null;
}

/**
 * For a counter entry, what countered the item, read from the server's
 * own line ("Counterspell countered Lightning Bolt"), which is already
 * redacted for this viewer. Null for any other entry.
 */
export function counteredBy(e: LogEvent): string | null {
  if (e.kind !== "counter") return null;
  const at = e.text.indexOf(" countered ");
  return at > 0 ? e.text.slice(0, at) : null;
}

/**
 * trackLinger folds one frame into the tracker: which of the previous
 * frame's stack items left, and what the log's new entries say about
 * each. `rects` is where each item was drawn in the previous frame. The
 * input tracker is not mutated.
 */
export function trackLinger(
  prev: LingerTracker,
  log: readonly LogEvent[] | undefined,
  items: readonly LingerItem[],
  opts: { reprime?: boolean; rects?: ReadonlyMap<string, Box> } = {},
): LingerFrame {
  const logged = track(prev.log, log, opts.reprime === true);
  const now = new Map(items.map((it) => [it.id, it]));
  const tracker: LingerTracker = { log: logged.tracker, items: now };
  if (logged.primed) return { tracker, lingers: [], fades: [], primed: true };

  const named = new Map<string, LogEvent>();
  for (const e of logged.fresh) {
    const id = entryItemID(e);
    if (id) named.set(id, e);
  }
  const lingers: Departure[] = [];
  const fades: Departure[] = [];
  for (const [id, item] of prev.items) {
    if (now.has(id)) continue;
    const rect = opts.rects?.get(id) ?? null;
    const e = named.get(id);
    const outcome = e ? (OUTCOME_OF[e.kind] ?? null) : null;
    if (e && outcome) {
      lingers.push({ item, outcome, by: counteredBy(e), seq: e.seq, rect });
    } else {
      fades.push({ item, outcome: null, by: null, seq: -1, rect });
    }
  }
  lingers.sort((a, b) => a.seq - b.seq);
  return { tracker, lingers, fades, primed: false };
}

/** The items the stack draws, as the linger keeps them. */
export function lingerItemsFrom(model: Pick<StackLaneModel, "stackItems">): LingerItem[] {
  return model.stackItems.map((it) => ({
    id: it.id,
    kind: it.raw.kind,
    name: it.name,
    // As the pile draws it: a card the viewer may not see is a back.
    image: it.previewCard ? cardImageURL(it.previewCard, "normal") : "/card-back.jpg",
    thumb: it.previewCard ? cardImageURL(it.previewCard, "small") : "/card-back-small.jpg",
    casterColor: it.casterColor,
    isCopy: it.raw.is_copy === true,
  }));
}

/**
 * How long each card of a burst lingers: 1.5 s shared out among the
 * `n` items that left in one frame, never under 400 ms.
 */
export function lingerDwellMs(n: number): number {
  return Math.max(LINGER_BURST_MIN_MS, LINGER_MS / Math.max(1, n));
}

export interface QueuedLinger extends Departure {
  dwellMs: number;
}

/**
 * The queue after `incoming` joins it: each new item gets its burst's
 * dwell, and at most LINGER_QUEUE_MAX are kept counting the one on
 * screen. The oldest waiting ones are dropped (the log keeps the
 * record); the one on screen is never cut short.
 */
export function enqueueLingers(
  pending: readonly QueuedLinger[],
  incoming: readonly Departure[],
  showing: boolean,
): QueuedLinger[] {
  const dwellMs = lingerDwellMs(incoming.length);
  const all = [...pending, ...incoming.map((d) => ({ ...d, dwellMs }))];
  const room = Math.max(0, LINGER_QUEUE_MAX - (showing ? 1 : 0));
  return all.slice(Math.max(0, all.length - room));
}

/** The badge for each outcome: its text and its tooltip. */
export const OUTCOME_BADGE: Record<LingerOutcome, { label: string; title: string }> = {
  resolved: { label: "Resolved", title: "resolved" },
  countered: { label: "Countered", title: "countered" },
  fizzled: {
    label: "Fizzled",
    title: "countered on resolution: no legal targets (CR 608.2b)",
  },
};

/** The announcer's line for a lingering item. */
export function lingerAnnouncement(d: Pick<Departure, "item" | "outcome" | "by">): string {
  switch (d.outcome) {
    case "resolved":
      return `${d.item.name} resolved`;
    case "countered":
      return d.by ? `${d.item.name} was countered by ${d.by}` : `${d.item.name} was countered`;
    case "fizzled":
      return `${d.item.name} fizzled: no legal targets`;
    default:
      return "";
  }
}

// ---- where it went -------------------------------------------------------

export type LingerZone = "battlefield" | "graveyard" | "exile" | "command";

export interface LingerDestination {
  zone: LingerZone;
  /** The card's instance ID, which is the spell's stack item ID. */
  cardID: string;
  /** The seat whose pile it went to: the owner, or a permanent's controller. */
  seat: string;
}

/**
 * Where a departed spell's card is now, or null when it should fade in
 * place: an ability (CR 608.2n, it "ceases to exist"), a copy (CR
 * 707.10), or a card in a zone this viewer cannot see. A spell's stack
 * item ID is its card's instance ID, and an instance ID survives a zone
 * move, so the card is found by it.
 */
export function lingerDestination(view: GameView, item: LingerItem): LingerDestination | null {
  if (item.kind !== "spell" || item.isCopy) return null;
  const id = item.id;
  const onField = view.battlefield?.cards.find((c) => c.instance_id === id);
  if (onField) return { zone: "battlefield", cardID: id, seat: onField.controller };
  for (const s of view.seats) {
    if (s.graveyard?.cards.some((c) => c.instance_id === id)) {
      return { zone: "graveyard", cardID: id, seat: s.id };
    }
  }
  const exiled = view.exile?.cards.find((c) => c.instance_id === id);
  if (exiled) return { zone: "exile", cardID: id, seat: exiled.owner };
  for (const s of view.seats) {
    if (s.command?.cards.some((c) => c.instance_id === id)) {
      return { zone: "command", cardID: id, seat: s.id };
    }
  }
  return null;
}

/** The PileBar attribute a flight lands on (ADR 0119 §3). */
export const PILE_ATTR = "data-pile";
export const PILE_OWNER_ATTR = "data-pile-owner";

function escapeAttr(s: string): string {
  return s.replace(/["\\]/g, "\\$&");
}

/** The selector for one seat's pile button. */
export function pileSelector(zone: LingerZone, seat: string): string {
  return `[${PILE_ATTR}="${zone}"][${PILE_OWNER_ATTR}="${escapeAttr(seat)}"]`;
}

/**
 * Where to look for the flight's landing, best first: the card itself
 * when the board draws it (a permanent's slot, the exile strip, the
 * command zone's card), then the owner's pile button. A graveyard card
 * goes to the pile even when it is the pile's top thumbnail.
 */
export function destinationSelectors(dest: LingerDestination): string[] {
  const pile = pileSelector(dest.zone, dest.seat);
  switch (dest.zone) {
    case "battlefield":
      return [cardSelector(dest.cardID)];
    case "graveyard":
      return [pile];
    case "exile":
    case "command":
      return [cardSelector(dest.cardID), pile];
  }
}

/** The flight's gate: the master switch and cardPlay, and never under reduced motion. */
export function flightAllowed(s: {
  animations: { enabled: boolean; cardPlay: boolean };
  accessibility: { reduceMotion: boolean };
}): boolean {
  return s.animations.enabled && s.animations.cardPlay && !s.accessibility.reduceMotion;
}

// ---- geometry ----------------------------------------------------------

function overlaps(a: Box, b: Box): boolean {
  return (
    a.left < b.left + b.width &&
    b.left < a.left + a.width &&
    a.top < b.top + b.height &&
    b.top < a.top + a.height
  );
}

/** The gap between a moved-out linger and the live stack. */
export const LINGER_GAP = 12;

/**
 * Where a lingering card is drawn. Where it was, unless that covers
 * what is live on the stack now (a new top card, the rest of the pile):
 * then out to the stack's right, or above it when the right has no
 * room, so it never covers the live top card.
 */
export function placeLinger(
  rect: Box,
  live: readonly Box[],
  board: { width: number; height: number },
): Box {
  if (!live.some((r) => overlaps(r, rect))) return rect;
  const right = Math.max(...live.map((r) => r.left + r.width));
  const left = right + LINGER_GAP;
  if (left + rect.width <= board.width - LINGER_GAP) return { ...rect, left };
  const top = Math.min(...live.map((r) => r.top)) - LINGER_GAP - rect.height;
  if (top >= LINGER_GAP) return { ...rect, top };
  return { ...rect, left: Math.max(LINGER_GAP, board.width - LINGER_GAP - rect.width) };
}

/** A row-shaped copy (a compact row, a ribbon cell, a peeking name line) or a card-shaped one. */
export function lingerShape(rect: Box): "card" | "row" {
  return rect.height >= rect.width * 0.9 ? "card" : "row";
}

/**
 * The FLIP step for the flight: the translation and scale, about the
 * copy's top-left corner, that put the copy's centre on the landing's
 * centre at the landing's size.
 */
export function flightTransform(from: Box, to: Box): { x: number; y: number; scale: number } {
  const fit =
    from.width > 0 && from.height > 0
      ? Math.min(to.width / from.width, to.height / from.height)
      : 1;
  const scale = Math.min(1.5, Math.max(0.1, Number.isFinite(fit) ? fit : 1));
  return {
    x: to.left + to.width / 2 - (from.left + (from.width * scale) / 2),
    y: to.top + to.height / 2 - (from.top + (from.height * scale) / 2),
    scale,
  };
}

// ---- the queue -----------------------------------------------------------

export interface LingerTimers {
  set(fn: () => void, ms: number): unknown;
  clear(handle: unknown): void;
}

const realTimers: LingerTimers = {
  set: (fn, ms) => setTimeout(fn, ms),
  clear: (h) => clearTimeout(h as ReturnType<typeof setTimeout>),
};

/**
 * LingerQueue shows departed cards one at a time, in resolution order:
 * `onShow` when one goes up, `onEnd` when its dwell is over (the shell
 * then flies or fades it). A new frame's departures join the queue
 * behind the one on screen, capped by enqueueLingers.
 */
export class LingerQueue {
  private showing: QueuedLinger | null = null;
  private pending: QueuedLinger[] = [];
  private timer: unknown = null;

  constructor(
    private readonly onShow: (l: QueuedLinger) => void,
    private readonly onEnd: (l: QueuedLinger) => void,
    private readonly timers: LingerTimers = realTimers,
  ) {}

  push(incoming: readonly Departure[]): void {
    if (incoming.length === 0) return;
    this.pending = enqueueLingers(this.pending, incoming, this.showing !== null);
    if (!this.showing) this.next();
  }

  /** The item on screen, if any. */
  get current(): QueuedLinger | null {
    return this.showing;
  }

  /** How many are waiting behind the one on screen. */
  get waiting(): number {
    return this.pending.length;
  }

  private next(): void {
    const up = this.pending.shift() ?? null;
    this.showing = up;
    if (!up) return;
    this.onShow(up);
    this.timer = this.timers.set(() => {
      this.timer = null;
      this.showing = null;
      this.onEnd(up);
      this.next();
    }, up.dwellMs);
  }

  /** Drop everything without ending it (a re-prime, an unmount). */
  clear(): void {
    if (this.timer !== null) this.timers.clear(this.timer);
    this.timer = null;
    this.showing = null;
    this.pending = [];
  }
}
