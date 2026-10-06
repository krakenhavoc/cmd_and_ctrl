// tutorialBoards.ts — the practice-table-shaped boards the tutorial's
// step tests build (tutorialSteps.test.ts), shared with the label guard
// (labels.test.ts, ADR 0125 §2.3), which walks every step's anchors and
// detours over them.

import type { StepContext } from "../tutorial";
import type { CardView, GameView } from "../protocol";

export const ME = "me";
export const BOT = "bot";

let n = 0;
export function card(name: string, type: string, extra: Partial<CardView> = {}): CardView {
  n += 1;
  return {
    instance_id: `${name.toLowerCase().replace(/\W+/g, "-")}-${n}`,
    name,
    owner: ME,
    controller: ME,
    type_line: type,
    ...extra,
  } as CardView;
}
export const forest = (extra: Partial<CardView> = {}) =>
  card("Forest", "Basic Land — Forest", {
    mana_abilities: [{ index: 0, label: "Add {G}" }] as unknown as CardView["mana_abilities"],
    ...extra,
  });
export const elves = (extra: Partial<CardView> = {}) =>
  card("Llanowar Elves", "Creature — Elf Druid", {
    mana_abilities: [{ index: 0, label: "Add {G}" }] as unknown as CardView["mana_abilities"],
    ...extra,
  });
export const walker = (extra: Partial<CardView> = {}) =>
  card("Phyrexian Walker", "Artifact Creature — Phyrexian Construct", extra);
export const marwyn = (extra: Partial<CardView> = {}) =>
  card("Marwyn, the Nurturer", "Legendary Creature — Elf Druid", extra);

/**
 * An open opening roll (ADR 0121 §3): the dice rolled so far in one
 * round of both seats, as [seat, result], and the chooser once there is
 * one.
 */
export interface Roll {
  rolls?: Array<[number, number]>;
  chooser?: number;
}

export interface Board {
  step?: string;
  active?: number;
  priority?: number;
  seq?: number;
  mine?: CardView[];
  theirs?: CardView[];
  hand?: CardView[];
  pool?: string[];
  landsPlayed?: number;
  stack?: CardView[];
  /** The viewer's command zone; empty by default. */
  command?: CardView[];
  /** The opening roll, open; absent by default. */
  roll?: Roll;
  /** The opening hands are open; `kept` is the viewer's. */
  mulligan?: { kept: boolean };
}

export function board(b: Board = {}): GameView {
  const zone = (kind: string, cards: CardView[] = []) => ({ kind, count: cards.length, cards });
  const seat = (id: string, i: number, hand: CardView[] = [], command: CardView[] = []) => ({
    id,
    name: id === ME ? "Player" : "Practice Bot",
    seat: i,
    life: 40,
    library: zone("library"),
    hand: zone("hand", hand),
    graveyard: zone("graveyard"),
    command: zone("command", command),
    hand_kept: id === ME && b.mulligan ? b.mulligan.kept : undefined,
    mana_pool: id === ME ? (b.pool ?? []) : [],
    lands_played_this_turn: id === ME ? (b.landsPlayed ?? 0) : 0,
  });
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, 0, b.hand ?? [forest(), elves()], b.command ?? []), seat(BOT, 1)],
    battlefield: zone("battlefield", [
      ...(b.mine ?? []),
      ...(b.theirs ?? []).map((c) => ({ ...c, owner: BOT, controller: BOT })),
    ]),
    stack: zone("stack", b.stack ?? []),
    exile: zone("exile"),
    stack_items: [],
    turn: {
      seq: b.seq ?? 1,
      number: 1,
      active_seat: b.active ?? 0,
      priority_holder: b.priority ?? b.active ?? 0,
      phase: "",
      step: b.step ?? "precombat_main",
    },
    mulligans_open: !!b.mulligan,
    ...(b.roll
      ? {
          opening_roll: {
            rounds: [
              {
                seats: [0, 1],
                rolls: (b.roll.rolls ?? []).map(([seat, result]) => ({ seat, result })),
              },
            ],
            ...(b.roll.chooser === undefined ? {} : { chooser: b.roll.chooser }),
          },
        }
      : {}),
  } as unknown as GameView;
}

export const ctx = (
  view: GameView | null,
  start: GameView | null = view,
  event = null,
  autopass = false,
): StepContext => ({
  view,
  start,
  viewerID: ME,
  event,
  client: { autopass },
});
export const auto = (view: GameView | null, start: GameView | null = view) =>
  ctx(view, start, null, true);
