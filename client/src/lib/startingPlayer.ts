import type { LogEvent, PlayerView } from "./protocol";

interface OpeningRollView {
  starting_seat?: number;
  seats: Array<Pick<PlayerView, "seat" | "name">>;
  log?: LogEvent[];
}

export interface OpeningRollWinner {
  seat: number;
  name: string;
  result?: number;
}

// openingRollWinner reads the durable public log rather than transient UI
// state, so a player who connects after Start sees the same roll and winner.
// Start rolls are the source-less d20 entries before the first turn event.
export function openingRollWinner(view: OpeningRollView | null): OpeningRollWinner | null {
  if (!view || view.starting_seat === undefined) return null;
  const player = view.seats.find((seat) => seat.seat === view.starting_seat);
  if (!player) return null;

  let result: number | undefined;
  for (const entry of view.log ?? []) {
    if (
      entry.kind === "roll" &&
      entry.seat === player.seat &&
      entry.sides === 20 &&
      !entry.card_id &&
      !entry.turn
    ) {
      result = entry.results?.at(-1);
    }
  }
  return { seat: player.seat, name: player.name, result };
}

export function openingRollText(winner: OpeningRollWinner): string {
  if (winner.result === undefined) return `${winner.name} goes first`;
  return `${winner.name} won the d20 roll with ${winner.result} and goes first`;
}
