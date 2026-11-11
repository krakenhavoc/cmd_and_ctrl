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
  /**
   * Set when this game began as a restart (CR 727.1a, #3856): the seat
   * is the player who restarted it, and `card` the card that did it
   * (Karn Liberated), when the log names one.
   */
  restarted?: { card?: string };
  /**
   * Set when this is a later game of a match (ADR 0222 §3, CR 103.1):
   * the seat is the player who chose, the loser of the game before,
   * `game` its number in the match and `drew` true when they chose to
   * draw (`chosen` is then the opponent, who plays first).
   */
  matchGame?: { game: number; drew: boolean };
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

  // CR 727.1a (#3856): a restarted game has no opening roll; the player
  // who restarted it goes first, and the log begins with that line.
  const restart = (view.log ?? []).find((entry) => entry.kind === "game_restarted");
  if (restart) {
    const name = nameOf(restart.seat);
    if (name === undefined) return null;
    return { seat: restart.seat, name, restarted: restart.label ? { card: restart.label } : {} };
  }

  // ADR 0222 §3: a later game of a match has no roll either. The loser
  // of the game before chose who plays first, and the log begins with
  // that line.
  const next = (view.log ?? []).find((entry) => entry.kind === "match_game");
  if (next) {
    const name = nameOf(next.seat);
    if (name === undefined) return null;
    const drew = next.label === "draw";
    const out: OpeningRollWinner = {
      seat: next.seat,
      name,
      matchGame: { game: next.amount ?? 0, drew },
    };
    const target = next.target_seat;
    if (target !== undefined && target !== next.seat) {
      const chosenName = nameOf(target);
      if (chosenName !== undefined) out.chosen = { seat: target, name: chosenName };
    }
    return out;
  }

  let choice: LogEvent | undefined;
  for (const entry of view.log ?? []) {
    if (entry.kind === "starting_player") choice = entry;
  }
  if (!choice) {
    // A table that never rolled (a replay or a restored game from
    // before the opening roll, ADR 0121; the practice table rolls too
    // since ADR 0125 §5.2): name who went first.
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
  if (winner.restarted) {
    const card = winner.restarted.card;
    return card
      ? `${winner.name} restarted the game with ${card} and goes first`
      : `${winner.name} restarted the game and goes first`;
  }
  if (winner.matchGame) {
    const game = winner.matchGame.game > 0 ? `Game ${winner.matchGame.game}: ` : "";
    return winner.matchGame.drew
      ? `${game}${winner.name} chose to draw`
      : `${game}${winner.name} chose to play first`;
  }
  const won =
    winner.result === undefined
      ? winner.name
      : `${winner.name} won the d20 roll with ${winner.result} and`;
  if (winner.chosen) return `${won} chose ${winner.chosen.name} to go first`;
  return `${won} goes first`;
}

// firstTurnHeadline is the opening hand's headline, as Arena puts it
// over the mulligan: who takes the first turn. The seat that goes first
// is the one chosen, else the winner, and the viewer's own seat reads
// "You go first". openingRollText says how it was decided, under it.
export function firstTurnHeadline(winner: OpeningRollWinner, viewerSeat?: number): string {
  const first = winner.chosen ?? winner;
  return first.seat === viewerSeat ? "You go first" : `${first.name} goes first`;
}
