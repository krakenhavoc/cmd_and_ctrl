import type { LogEvent, OpeningRollView, PlayerView } from "./protocol";

interface StartingView {
  starting_seat?: number;
  opening_roll?: OpeningRollView;
  seats: Array<Pick<PlayerView, "seat" | "name">>;
  log?: LogEvent[];
}

export interface OpeningRollWinner {
  /** The seat that won the opening roll and chose. */
  seat: number;
  name: string;
  /** The winning roll, when the log still has it. */
  result?: number;
  /** The seat chosen to take the first turn, when it is not the winner's. */
  chosen?: { seat: number; name: string };
}

// openingRollWinner reads the durable public log rather than transient UI
// state, so a player who connects after the choice sees the same result.
// The `starting_player` entry names the winner (`seat`) and the seat they
// chose (`target_seat`); the `opening_roll` entry labelled "won" carries
// the winning roll (ADR 0121 §3). Nothing is guessed from bare d20s, so a
// d20 rolled at the table before turn 1 is never taken for the opening
// roll. While the roll is still open there is no winner yet, and
// `starting_seat` (which then reads 0) is ignored.
export function openingRollWinner(view: StartingView | null): OpeningRollWinner | null {
  if (!view || view.opening_roll) return null;
  const nameOf = (seat: number) => view.seats.find((s) => s.seat === seat)?.name;

  let choice: LogEvent | undefined;
  for (const entry of view.log ?? []) {
    if (entry.kind === "starting_player") choice = entry;
  }
  if (!choice) {
    // A table that never rolled (the practice table, ADR 0076) or a
    // replay from before the opening roll: name who went first.
    if (view.starting_seat === undefined) return null;
    const name = nameOf(view.starting_seat);
    return name === undefined ? null : { seat: view.starting_seat, name };
  }

  const name = nameOf(choice.seat);
  if (name === undefined) return null;
  let result: number | undefined;
  for (const entry of view.log ?? []) {
    if (
      entry.kind === "opening_roll" &&
      entry.label === "won" &&
      entry.seat === choice.seat &&
      entry.seq < choice.seq
    ) {
      result = entry.results?.[0];
    }
  }
  const out: OpeningRollWinner = { seat: choice.seat, name };
  if (result !== undefined) out.result = result;
  const target = choice.target_seat;
  if (target !== undefined && target !== choice.seat) {
    const chosenName = nameOf(target);
    if (chosenName !== undefined) out.chosen = { seat: target, name: chosenName };
  }
  return out;
}

// openingRollText is the `.opening-roll` pill and the mulligan sheet's
// line: "Carol won the d20 roll with 20 and goes first", or "… and chose
// Bob to go first" when the winner gave the first turn away.
export function openingRollText(winner: OpeningRollWinner): string {
  const won =
    winner.result === undefined
      ? winner.name
      : `${winner.name} won the d20 roll with ${winner.result} and`;
  if (winner.chosen) return `${won} chose ${winner.chosen.name} to go first`;
  return `${won} goes first`;
}
