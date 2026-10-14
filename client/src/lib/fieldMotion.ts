// fieldMotion — permanents arriving and leaving where you can see it.
//
// The stack linger (ADR 0119 §3) shows a spell coming off the stack and
// flying to where it went, and the combat strikes (ADR 0134) crumble a
// creature that died in combat. Everything else that changes a board
// used to happen between two frames with nothing to watch: a creature
// destroyed by a spell, a sacrificed permanent, a bounce, an exile, a
// land drop, a token, a reanimated creature, a stolen permanent. On a
// four-player table the change is one small tile on someone else's
// board, and it is easy to miss entirely.
//
// So two cues:
//
//   - A DEPARTURE. A permanent that left the battlefield is drawn again
//     where it stood, lifts, and flies to where it went: its owner's
//     graveyard, exile or library pile, its card in the exile strip,
//     command zone or your hand, or the owner's seat when it went to a
//     hand you cannot see. A permanent that changed controller flies to
//     its new slot. A token, or a card that went nowhere this viewer
//     can find, fades in place.
//   - An ARRIVAL. A permanent that is new on a board you do not control
//     is ringed for a few seconds, so the eye finds it after the fact.
//
// DISPLAY ONLY, as the linger is: the board already shows the frame's
// end state and input stays live under the layer.
//
// THE SIGNAL is the battlefield itself, frame to frame, by instance ID
// (an instance ID survives a zone move). Where a card went is read from
// the next frame when the card is in a zone this viewer can see, and
// from the public log's `zone` entry when it is not (a hand or a
// library: the server names the card because its old zone was public).
//
// PRIMING. The first frame after mounting, a reconnect, a replay toggle
// or an undo moves nothing: the log watermark from combatBeats.ts
// decides, so "new" means the same here as for the combat beats and
// the linger.
//
// Plain functions over plain data, so all of it is unit-testable
// without mounting a component (#689). FieldMotion.svelte measures,
// draws and flies.

import type { GameView, LogEvent } from "./protocol";
import { cardSelector, seatSelector } from "./boardAnchor";
import { emptyBeatTracker, track, type BeatTracker } from "./combatBeats";
import { phasedOutCards } from "./phasedOut";
import { pileSelector, type Box } from "./stackLinger";

/** How long a departed permanent holds where it stood before it flies. Reading time, never speed-scaled. */
export const DEPART_HOLD_MS = 260;
/** The flight to where it went, before animations.speed. */
export const DEPART_FLIGHT_MS = 460;
/** A permanent with nowhere to fly to fades in this long. */
export const DEPART_FADE_MS = 320;
/** At most this many permanents fly at once; the rest of a board wipe fades. */
export const DEPART_FLIGHT_MAX = 16;
/** How long an arrival stays ringed. Reading time, never speed-scaled. */
export const ARRIVAL_MS = 2600;

/** One permanent as the board drew it, kept so it can be drawn again after it has gone. */
export interface FieldCard {
  id: string;
  owner: string;
  controller: string;
  isToken: boolean;
  tapped: boolean;
  /** The image the tile showed (its own `src`), or null for none. */
  image: string | null;
  /** Where the tile was drawn, in board pixels. */
  rect: Box;
}

export type FieldZone = "graveyard" | "exile" | "library" | "hand" | "command" | "battlefield";

/** Where a departed permanent went, as far as this viewer can tell. */
export interface FieldDestination {
  zone: FieldZone;
  /** The seat whose pile, hand or board it went to. */
  seat: string;
  /** True when the card itself is drawn there (an exile strip card, your hand, a new slot). */
  cardVisible: boolean;
}

export interface FieldDeparture {
  card: FieldCard;
  /** Null: fade in place. */
  dest: FieldDestination | null;
}

export interface FieldTracker {
  log: BeatTracker;
  /** The previous frame's permanents: instance ID → controller. */
  field: Map<string, string>;
}

export function emptyFieldTracker(): FieldTracker {
  return { log: emptyBeatTracker(), field: new Map() };
}

export interface FieldFrame {
  tracker: FieldTracker;
  departures: FieldDeparture[];
  /** IDs of permanents new on a board this viewer does not control. */
  arrivals: string[];
  /** True when this frame primed or rewound: nothing moves. */
  quiet: boolean;
}

/** Every permanent the board draws, phased-out ones included (Board.svelte puts those back in their row). */
export function fieldControllers(view: GameView): Map<string, string> {
  const out = new Map<string, string>();
  for (const c of view.battlefield?.cards ?? []) out.set(c.instance_id, c.controller);
  for (const c of phasedOutCards(view)) out.set(c.instance_id, c.controller);
  return out;
}

const ZONES: readonly FieldZone[] = ["graveyard", "exile", "library", "hand", "command"];

function asZone(z: string | undefined): FieldZone | null {
  return z && (ZONES as readonly string[]).includes(z) ? (z as FieldZone) : null;
}

/**
 * Where a permanent that left the battlefield is now. The next frame
 * first: a graveyard, the exile strip, a command zone or the viewer's
 * own hand, wherever the card is drawn by its instance ID. Then the
 * log's `zone` entry for a card that went somewhere hidden. Null when
 * neither says: the copy fades in place.
 */
export function fieldDestination(
  view: GameView,
  card: Pick<FieldCard, "id" | "owner">,
  fresh: readonly LogEvent[],
): FieldDestination | null {
  const id = card.id;
  for (const s of view.seats ?? []) {
    if (s.graveyard?.cards.some((c) => c.instance_id === id)) {
      return { zone: "graveyard", seat: s.id, cardVisible: false };
    }
  }
  const exiled = view.exile?.cards.find((c) => c.instance_id === id);
  if (exiled) return { zone: "exile", seat: exiled.owner, cardVisible: true };
  for (const s of view.seats ?? []) {
    if (s.command?.cards.some((c) => c.instance_id === id)) {
      return { zone: "command", seat: s.id, cardVisible: true };
    }
    if (s.hand?.cards.some((c) => c.instance_id === id)) {
      return { zone: "hand", seat: s.id, cardVisible: true };
    }
  }
  // The newest entry wins: a card can pass through a zone in one frame.
  for (let i = fresh.length - 1; i >= 0; i--) {
    const e = fresh[i];
    if (e.kind !== "zone" || e.card_id !== id || e.old_zone !== "battlefield") continue;
    const zone = asZone(e.new_zone);
    if (!zone) return null;
    return { zone, seat: card.owner, cardVisible: false };
  }
  return null;
}

/**
 * The cards a fresh combat damage entry names, as source or as what it
 * hit. The combat strikes already show those (ADR 0134: a creature that
 * dies in combat crumbles), so a departure would draw it twice.
 */
export function combatCards(fresh: readonly LogEvent[]): Set<string> {
  const out = new Set<string>();
  for (const e of fresh) {
    if (e.kind !== "damage" || e.combat !== true) continue;
    if (e.card_id) out.add(e.card_id);
    if (e.target) out.add(e.target);
  }
  return out;
}

/**
 * trackField folds one frame into the tracker: which of the previous
 * frame's permanents left (or changed controller), where each went, and
 * which permanents are new on someone else's board. `drawn` is what the
 * board drew last frame, by instance ID. The input tracker is not
 * mutated.
 */
export function trackField(
  prev: FieldTracker,
  view: GameView,
  viewerID: string | null,
  drawn: ReadonlyMap<string, FieldCard>,
  opts: { reprime?: boolean } = {},
): FieldFrame {
  const logged = track(prev.log, view.log, opts.reprime === true);
  const field = fieldControllers(view);
  const tracker: FieldTracker = { log: logged.tracker, field };
  if (logged.primed || logged.rewound) {
    return { tracker, departures: [], arrivals: [], quiet: true };
  }

  const skip = combatCards(logged.fresh);
  const departures: FieldDeparture[] = [];
  for (const [id, controller] of prev.field) {
    const now = field.get(id);
    if (now === controller) continue;
    const card = drawn.get(id);
    if (!card || skip.has(id)) continue;
    if (now !== undefined) {
      // Still on the battlefield under someone else: fly to its new slot.
      departures.push({ card, dest: { zone: "battlefield", seat: now, cardVisible: true } });
      continue;
    }
    // CR 111.7: a token that leaves the battlefield ceases to exist.
    const dest = card.isToken ? null : fieldDestination(view, card, logged.fresh);
    departures.push({ card, dest });
  }

  const arrivals: string[] = [];
  for (const [id, controller] of field) {
    if (prev.field.get(id) === controller) continue;
    if (viewerID !== null && controller === viewerID) continue;
    arrivals.push(id);
  }
  return { tracker, departures, arrivals, quiet: false };
}

/**
 * Where to look for a departure's landing, best first: the card itself
 * when it is drawn there, then the owner's pile, then the owner's seat
 * (a hand or a command zone with no pile of its own).
 */
export function fieldLandingSelectors(dest: FieldDestination, cardID: string): string[] {
  const out: string[] = [];
  if (dest.cardVisible) out.push(cardSelector(cardID));
  switch (dest.zone) {
    case "graveyard":
    case "exile":
    case "library":
      out.push(pileSelector(dest.zone, dest.seat));
      break;
    case "hand":
    case "command":
      out.push(seatSelector(dest.seat));
      break;
    case "battlefield":
      break;
  }
  return out;
}

/** A plain-words line for what happened, for the copy's caption. */
export function departureCaption(d: FieldDeparture): string {
  if (!d.dest) return d.card.isToken ? "gone" : "left";
  switch (d.dest.zone) {
    case "graveyard":
      return "to graveyard";
    case "exile":
      return "exiled";
    case "library":
      return "to library";
    case "hand":
      return "to hand";
    case "command":
      return "to command zone";
    case "battlefield":
      return "changed control";
  }
}

/**
 * The upright box a tile's copy is drawn in. A tapped tile's measured
 * box is its rotated bounding box; the copy is drawn upright at the
 * tile's own size, about the same centre, and turned by CSS.
 */
export function uprightBox(card: Pick<FieldCard, "rect" | "tapped">): Box {
  const r = card.rect;
  if (!card.tapped || r.width <= r.height) return r;
  const cx = r.left + r.width / 2;
  const cy = r.top + r.height / 2;
  return { left: cx - r.height / 2, top: cy - r.width / 2, width: r.height, height: r.width };
}

/**
 * ArrivalClock keeps each ringed permanent's expiry. `ring` adds or
 * restarts one; `live` is the set still ringed at `now`, and drops the
 * rest.
 */
export class ArrivalClock {
  private until = new Map<string, number>();

  ring(ids: readonly string[], now: number): void {
    for (const id of ids) this.until.set(id, now + ARRIVAL_MS);
  }

  live(now: number): string[] {
    const out: string[] = [];
    for (const [id, t] of this.until) {
      if (t > now) out.push(id);
      else this.until.delete(id);
    }
    return out;
  }

  /** The soonest expiry still pending, or null when nothing is ringed. */
  next(): number | null {
    let soonest: number | null = null;
    for (const t of this.until.values()) if (soonest === null || t < soonest) soonest = t;
    return soonest;
  }

  clear(): void {
    this.until.clear();
  }
}
