// openingRoll — the table during the opening roll (ADR 0121 §6).
//
// While `GameView.opening_roll` is present, every seat rolls a d20, the
// tied leaders roll again, and the one leader left chooses who takes the
// first turn (CR 103.1). Nothing is dealt until then. This module turns
// the view into what the screen draws, as pure functions:
//
//   - the strip's `opening roll` banner: one chip per seat, with its
//     result, "rolling…" or "—" (not in this round), the round's tied
//     leaders outlined and the chooser marked once known;
//   - what the action dock asks the viewer: Roll, the host's "Roll for
//     everyone left", the chooser's sheet, or nothing but a status line;
//   - which `roll` log entries are opening dice, so the strip raises no
//     cue for them (the banner already shows each one) and the dice
//     layer keeps them settled by their seats until the choice.

import type { GameView, LogEvent, OpeningRollView, PlayerView } from "./protocol";

/**
 * isOpeningDieLog reports whether a log entry is one opening d20. Each
 * opening die is a `roll` line with no card behind it, written before
 * the first turn began (ADR 0121 §3). A card's roll always names its
 * card, and a table roll (§5) is its own `table_roll` kind, so neither
 * is ever taken for one.
 */
export function isOpeningDieLog(log: LogEvent): boolean {
  return log.kind === "roll" && !log.card_id && !log.turn && log.sides === 20;
}

export type OpeningChipState = "rolled" | "rolling" | "out";

export interface OpeningChip {
  seat: number;
  name: string;
  /** rolled: has a die this round; rolling: owes one; out: not in this round. */
  state: OpeningChipState;
  result?: number;
  /** A tied leader of the round before, rolling again in this one. */
  tied: boolean;
  /** The seat that chooses who takes the first turn. */
  chooser: boolean;
}

export interface OpeningRollModel {
  /** 0 for the first round; each later round is a reroll of a tie. */
  roundIndex: number;
  /** The seats rolling in the current round. */
  roundSeats: number[];
  /** Seats in the current round that have not rolled yet, in seat order. */
  waiting: number[];
  /** The chooser's seat, once one leader is left. */
  chooser: number | null;
  /** In a reroll: the result the current round's seats tied on. */
  tieResult: number | null;
  /** Once there is a chooser: the roll that won it. */
  winResult: number | null;
  /** One chip per seat, in seat order. */
  chips: OpeningChip[];
}

type SeatLike = Pick<PlayerView, "seat" | "name"> & { eliminated?: boolean };

function maxResult(or: OpeningRollView, round: number, seats: readonly number[]): number | null {
  let best: number | null = null;
  for (const d of or.rounds[round]?.rolls ?? []) {
    if (seats.includes(d.seat) && (best === null || d.result > best)) best = d.result;
  }
  return best;
}

/** The model of an open opening roll, or null when none is open. */
export function openingRollModel(
  view: Pick<GameView, "opening_roll"> & { seats: SeatLike[] },
): OpeningRollModel | null {
  const or = view.opening_roll;
  if (!or || or.rounds.length === 0) return null;
  const roundIndex = or.rounds.length - 1;
  const round = or.rounds[roundIndex];
  const chooser = or.chooser ?? null;
  const rolled = new Map(round.rolls.map((d) => [d.seat, d.result]));
  const waiting =
    chooser === null ? round.seats.filter((s) => !rolled.has(s)).sort((a, b) => a - b) : [];
  const tieResult = roundIndex > 0 ? maxResult(or, roundIndex - 1, round.seats) : null;
  const winResult = chooser !== null ? (rolled.get(chooser) ?? null) : null;

  const chips: OpeningChip[] = [...view.seats]
    .sort((a, b) => a.seat - b.seat)
    .map((s) => {
      const inRound = round.seats.includes(s.seat) && !s.eliminated;
      const result = inRound ? rolled.get(s.seat) : undefined;
      const state: OpeningChipState = !inRound
        ? "out"
        : result !== undefined
          ? "rolled"
          : "rolling";
      return {
        seat: s.seat,
        name: s.name,
        state,
        ...(result !== undefined ? { result } : {}),
        tied: chooser === null && roundIndex > 0 && inRound,
        chooser: chooser === s.seat,
      };
    });

  return {
    roundIndex,
    roundSeats: [...round.seats],
    waiting,
    chooser,
    tieResult,
    winResult,
    chips,
  };
}

/** The text a chip shows for its result. */
export function openingChipText(chip: OpeningChip): string {
  if (chip.state === "rolled") return String(chip.result);
  if (chip.state === "rolling") return "rolling…";
  return "—";
}

/** "Alice", "Alice and Bob", "Alice, Bob and Carol". */
export function nameList(names: readonly string[]): string {
  if (names.length <= 1) return names[0] ?? "";
  return `${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}`;
}

function nameOf(seats: readonly SeatLike[], seat: number): string {
  return seats.find((s) => s.seat === seat)?.name ?? `Seat ${seat + 1}`;
}

/** The question line on a seat's Roll request (ADR 0121 §6). */
export function rollQuestion(model: OpeningRollModel): string {
  if (model.roundIndex > 0 && model.tieResult !== null) {
    return `You tied with ${model.tieResult}. Roll again.`;
  }
  return "Roll a d20. The highest roll chooses who goes first.";
}

/** "Waiting for Bob and Dave to roll", or "" when nobody is left. */
export function waitingText(model: OpeningRollModel, seats: readonly SeatLike[]): string {
  if (model.waiting.length === 0) return "";
  return `Waiting for ${nameList(model.waiting.map((s) => nameOf(seats, s)))} to roll`;
}

/**
 * What the action dock asks the viewer during the opening roll:
 *
 *   - roll: the viewer owes a die in this round. The request's primary
 *     is Roll; the host also gets "Roll for everyone left" while some
 *     other seat has not rolled.
 *   - wait: the host has nothing to roll, others do. The request reads
 *     "Waiting for … to roll" and offers only "Roll for everyone left".
 *   - choose: the viewer won and chooses who takes the first turn.
 *   - none: nothing to answer; the dock's status line says why.
 */
export type OpeningRollAsk =
  | { kind: "roll"; question: string; others: number[] }
  | { kind: "wait"; question: string; others: number[] }
  | { kind: "choose"; result: number | null }
  | { kind: "none" };

export function openingRollAsk(
  model: OpeningRollModel | null,
  seats: readonly SeatLike[],
  viewerSeat: number | null,
  isHost: boolean,
): OpeningRollAsk {
  if (!model || viewerSeat === null) return { kind: "none" };
  if (model.chooser !== null) {
    return model.chooser === viewerSeat
      ? { kind: "choose", result: model.winResult }
      : { kind: "none" };
  }
  const others = model.waiting.filter((s) => s !== viewerSeat);
  if (model.waiting.includes(viewerSeat)) {
    return { kind: "roll", question: rollQuestion(model), others: isHost ? others : [] };
  }
  if (isHost && others.length > 0) {
    return { kind: "wait", question: waitingText(model, seats), others };
  }
  return { kind: "none" };
}

/**
 * The dock's status line for a viewer with nothing to answer: who the
 * table is waiting for. "Carol is choosing who goes first" once there is
 * a chooser, "Waiting for Bob and Dave to roll" before.
 */
export function openingRollStatus(
  view: Pick<GameView, "opening_roll"> & { seats: SeatLike[] },
): string {
  const model = openingRollModel(view);
  if (!model) return "";
  if (model.chooser !== null) {
    return `${nameOf(view.seats, model.chooser)} is choosing who goes first`;
  }
  return waitingText(model, view.seats);
}

export interface StartingChoice {
  seat: number;
  /** The button's name: "I go first" for the chooser's own seat. */
  label: string;
  name: string;
  self: boolean;
}

/**
 * The chooser's buttons, one per seat still in the game, in turn order
 * (seat order): "<name> goes first", or "I go first" for their own.
 */
export function startingChoices(seats: readonly SeatLike[], chooser: number): StartingChoice[] {
  return [...seats]
    .filter((s) => !s.eliminated)
    .sort((a, b) => a.seat - b.seat)
    .map((s) => {
      const self = s.seat === chooser;
      return {
        seat: s.seat,
        name: s.name,
        self,
        label: self ? "I go first" : `${s.name} goes first`,
      };
    });
}

/** The confirm's question when the chooser gives the first turn away (owner decision 6). */
export function giveAwayQuestion(name: string): string {
  return `Let ${name} take the first turn?`;
}

/** The chooser sheet's lead line. */
export function chooserLead(result: number | null): string {
  return result === null
    ? "You won the roll. Choose who takes the first turn."
    : `You won the roll with ${result}. Choose who takes the first turn.`;
}

/**
 * The latest opening die of each seat in the log window, in seat order:
 * the standings the dice layer keeps settled by each seat until the
 * choice. A seat that rerolled shows its reroll.
 */
export function latestOpeningDice(logs: readonly LogEvent[] | undefined): LogEvent[] {
  const bySeat = new Map<number, LogEvent>();
  for (const log of logs ?? []) {
    if (isOpeningDieLog(log) && log.seat >= 0) bySeat.set(log.seat, log);
  }
  return [...bySeat.entries()].sort(([a], [b]) => a - b).map(([, log]) => log);
}
