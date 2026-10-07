// tableWhen.ts — the small questions the table hints' `when` ask of the
// game view (ADR 0125 §3.7). Pure, so each is unit-tested on its own.
// A hint reads the HintContext and nothing else (§3.1): these take it.

import type { CardView, PlayerView } from "../protocol";
import type { HintContext } from "./hint";

/** viewerSeat is the viewer's own seat, or null for a spectator. */
export function viewerSeat(c: HintContext): PlayerView | null {
  if (!c.view || !c.viewerID) return null;
  return c.view.seats.find((s) => s.id === c.viewerID) ?? null;
}

/** stackIsEmpty reports whether the stack holds no item (and none is waiting). */
export function stackIsEmpty(c: HintContext): boolean {
  return (c.view?.stack_items ?? []).length === 0 && (c.view?.stack.count ?? 0) === 0;
}

/**
 * atViewersSecondTurn reports whether the viewer's second turn has
 * begun. `turn.number` is the table's rotation, 1-indexed, so the
 * viewer's nth turn is round n, reached once the seats before them in
 * the order have gone.
 */
export function atViewersSecondTurn(c: HintContext): boolean {
  const me = viewerSeat(c);
  if (!me || !c.view) return false;
  const { number, active_seat } = c.view.turn;
  return number > 2 || (number === 2 && active_seat >= me.seat);
}

/** firstOpponent is the first other seat still in the game, in seat order. */
export function firstOpponent(c: HintContext): PlayerView | null {
  const me = viewerSeat(c);
  if (!me || !c.view) return null;
  return (
    [...c.view.seats]
      .sort((a, b) => a.seat - b.seat)
      .find((s) => s.id !== me.id && !s.eliminated) ?? null
  );
}

/** controlsAbility reports whether a battlefield card is the viewer's and offers an ability. */
export function controlsAbility(c: HintContext): boolean {
  const cards: CardView[] = c.view?.battlefield.cards ?? [];
  return cards.some(
    (card) => card.controller === c.viewerID && (card.activated_abilities?.length ?? 0) > 0,
  );
}

/**
 * EXERT_ROW is the static ability row the server draws for "You may
 * exert this creature as it attacks" (server/internal/game
 * ability_rows.go). The server's own words, not oracle text.
 */
export const EXERT_ROW = "You may exert this creature as it attacks.";

/** controlsExertCreature reports whether the viewer controls a creature that may be exerted as it attacks (ADR 0130). */
export function controlsExertCreature(c: HintContext): boolean {
  const cards: CardView[] = c.view?.battlefield.cards ?? [];
  return cards.some(
    (card) =>
      card.controller === c.viewerID &&
      (card.ability_rows ?? []).some((r) => r.kind === "static" && r.label === EXERT_ROW),
  );
}

/** hasHoverPointer reports a fine pointer that can rest on things (a mouse, a trackpad). */
export function hasHoverPointer(): boolean {
  if (typeof matchMedia !== "function") return false;
  return matchMedia("(hover: hover) and (pointer: fine)").matches;
}
