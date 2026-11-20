// table.ts — the fake table the fx preview harness draws (#4868).
//
// A scenario describes the table before and after one moment as a list
// of cards (FxCard), plus the log entries the server would send for it.
// toView turns a table into the GameView the real effect layers read, and
// Stage.svelte draws the same cards as a board with the data attributes
// the layers measure (data-instance-id, data-seat-id, data-pile,
// data-hand-owner, data-stack-item-id). Nothing here is sent anywhere.

import { cardImageURL } from "../../src/lib/cardImage";
import type { CardView, GameView, LogEvent } from "../../src/lib/protocol";

/** The four seats. Seat 0 is the viewer. */
export const SEATS = [
  { id: "p0", name: "You", color: "#ff7a1a" },
  { id: "p1", name: "Bea", color: "#6fa8ff" },
  { id: "p2", name: "Cal", color: "#6fe3a4" },
  { id: "p3", name: "Dee", color: "#b08aff" },
] as const;

export type Seat = 0 | 1 | 2 | 3;

export type Zone = "battlefield" | "stack" | "command" | "graveyard" | "hand" | "exile";

/** One card on the fake table. Only `id` is required. */
export interface FxCard {
  id: string;
  /** Shown on its placeholder art. Defaults to the id. */
  name?: string;
  /** Owner and controller. Defaults to seat 1, the opponent across the table. */
  seat?: Seat;
  /** Defaults to the battlefield. */
  zone?: Zone;
  /** Defaults to "Creature — Ogre". */
  type?: string;
  token?: boolean;
  commander?: boolean;
  tapped?: boolean;
  /** Draws a power/toughness badge on the tile, e.g. "2/2". */
  pt?: string;
  /** Draws a counter pip on the tile, of this kind ("charge"). */
  counter?: string;
  /** Draws the tile tucked under this permanent (an Equipment or an Aura). */
  attachedTo?: string;
  /** A double-faced card, with this face up (0 the front, 1 the back). */
  face?: number;
  /** On the stack: its stack item's ID. Defaults to the card's id. */
  item?: string;
}

/** A seat that has left the game in this table, and why (the log's `cause`). */
export interface FxOut {
  seat: Seat;
  cause: string;
}

/** A log entry as a scenario writes it: the harness stamps `seq` and `turn`. */
export type Entry = Omit<LogEvent, "seq" | "seat" | "text"> & { seat?: number; text?: string };

export const seatOf = (c: FxCard): Seat => c.seat ?? 1;
export const zoneOf = (c: FxCard): Zone => c.zone ?? "battlefield";

/** The placeholder art's URL: the harness's dev server draws it (vite.config.ts). */
export function artOf(c: FxCard, size: "small" | "normal" = "small"): string {
  return cardImageURL(cardView(c), size) ?? "/card-back-small.jpg";
}

export function cardView(c: FxCard): CardView {
  const owner = SEATS[seatOf(c)].id;
  const name = c.name ?? c.id;
  const dfc = c.face !== undefined;
  return {
    instance_id: c.id,
    name,
    owner,
    controller: owner,
    known_by_you: true,
    type_line: c.type ?? "Creature — Ogre",
    scryfall_id: `fx-${encodeURIComponent(name)}`,
    is_token: c.token || undefined,
    is_commander: c.commander || undefined,
    tapped: c.tapped || undefined,
    attached_to: c.attachedTo,
    ...(dfc
      ? {
          layout: "transform",
          active_face: c.face,
          faces: [
            { name: `${name} (front)`, image: `/cards/fx-${encodeURIComponent(name)}/image` },
            { name: `${name} (back)`, image: `/cards/fx-${encodeURIComponent(name)}/image?face=1` },
          ],
        }
      : {}),
  } as unknown as CardView;
}

/** The cards of one zone, of one seat (or of every seat). */
export function cardsIn(cards: readonly FxCard[], zone: Zone, seat?: Seat): FxCard[] {
  return cards.filter((c) => zoneOf(c) === zone && (seat === undefined || seatOf(c) === seat));
}

const pile = (
  kind: string,
  owner: string | undefined,
  cards: CardView[],
  count = cards.length,
) => ({
  kind,
  owner,
  count,
  cards,
});

/** The GameView the effect layers read for this table and log. */
export function toView(cards: readonly FxCard[], log: LogEvent[], out: FxOut[] = []): GameView {
  const views = (zone: Zone, seat?: Seat) => cardsIn(cards, zone, seat).map(cardView);
  const seats = SEATS.map((s, i) => {
    const n = i as Seat;
    const gone = out.find((o) => o.seat === n);
    const hand = views("hand", n);
    return {
      id: s.id,
      name: s.name,
      seat: n,
      life: 40,
      library: pile("library", s.id, [], 60),
      // The viewer sees their own hand; everyone else's is a count.
      hand: pile("hand", s.id, n === 0 ? hand : [], gone ? 0 : hand.length),
      graveyard: pile("graveyard", s.id, views("graveyard", n)),
      command: pile("command", s.id, views("command", n)),
      commander_damage: {},
      life_history: [],
      mana_pool: [],
      ...(gone ? { eliminated: true, loss_cause: gone.cause } : {}),
    };
  });
  return {
    id: "fx-preview",
    state: "active",
    seats,
    battlefield: pile("battlefield", undefined, views("battlefield")),
    stack: pile("stack", undefined, views("stack")),
    exile: pile("exile", undefined, views("exile")),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 1,
      number: 3,
      active_seat: 1,
      priority_holder: 1,
      phase: "precombat_main",
      step: "precombat_main",
    },
    mulligans_open: false,
    log,
  } as unknown as GameView;
}

/** Stamps a scenario's entries as one frame's new log: seq from 1, turn 3. */
export function stamp(entries: readonly Entry[]): LogEvent[] {
  return entries.map((e, i) => ({ turn: 3, seat: -1, text: "", ...e, seq: i + 1 }) as LogEvent);
}
