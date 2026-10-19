// outsideRight.ts — what the action dock offers while an effect lets the
// viewer reach the cards they own outside the game (ADR 0191, amended
// 2026-10-19 for #3833 and #4054).
//
// The sideboard is not on the table: it is browsed on the Decks page, and
// opponents never see it. In a game a card outside the game is reached
// only through an effect, and two of those effects leave a standing right
// that needs a place to be used from:
//
//   - Wish with the card database as its pool (ADR 0198's #3955
//     amendment): "you may play a card you own from outside the game this
//     turn". The card is named when it is played, so the right offers a
//     "Name a card to play" button (`outside_play.can_name`).
//   - A card in the viewer's outside-the-game container a live permission
//     opens (Spawnsire of Ulamog's free casts, Wish with a sideboard as its
//     pool). The server stamps `castable_here` on exactly those, and the
//     dock offers a button that opens the cards the right lets them cast
//     or play, each with its own cast button.
//
// NOTHING HERE IS A RULE. Both inputs are server answers on the viewer's
// own frame, which carries them to the viewer alone.

import type { GameView } from "./protocol";
import { castableFromZone } from "./zoneBrowser.logic";

/** One button the dock shows: name a card, or open the cards a right opens. */
export interface OutsideChip {
  kind: "name" | "open";
  label: string;
  title: string;
}

/** The cards in the viewer's outside-the-game container a live right opens. */
export function outsideCastable(view: GameView | null | undefined, viewerID: string | null) {
  if (!view || !viewerID) return [];
  const seat = view.seats?.find((s) => s.id === viewerID);
  return (seat?.sideboard ?? []).filter((c) => castableFromZone(c, "sideboard"));
}

/**
 * outsideChips is the dock's buttons for the viewer's live rights over
 * the cards they own outside the game, none when there is no such right.
 * Only the viewer's own seat is read.
 */
export function outsideChips(
  view: GameView | null | undefined,
  viewerID: string | null,
): OutsideChip[] {
  if (!view || !viewerID) return [];
  const seat = view.seats?.find((s) => s.id === viewerID);
  if (!seat) return [];
  const out: OutsideChip[] = [];
  const right = seat.outside_play;
  if (right?.can_name) {
    out.push({
      kind: "name",
      label: "name a card to play",
      title: `${right.source_name}: pick any card from the card database; it is played now, at its own timing and cost`,
    });
  }
  const castable = outsideCastable(view, viewerID);
  if (castable.length > 0) {
    out.push({
      kind: "open",
      label: `outside the game (${castable.length})`,
      title: `Cards you own outside the game that ${
        right?.source_name ?? "an effect"
      } lets you cast or play`,
    });
  }
  return out;
}

/** The dock's `outside` prop: the buttons and what pressing each does. */
export interface OutsideDock {
  chips: OutsideChip[];
  onPress: (kind: OutsideChip["kind"]) => void;
}

/**
 * outsideDock is the action dock's prop for the viewer's live rights over
 * the cards they own outside the game, or undefined when there is none.
 * `name` sends the `outside_play` action (Wish's naming pick); `open`
 * shows the cards the right lets the viewer cast or play.
 */
export function outsideDock(
  view: GameView | null | undefined,
  viewerID: string | null,
  actions: { name: () => void; open: () => void },
): OutsideDock | undefined {
  const chips = outsideChips(view, viewerID);
  if (chips.length === 0) return undefined;
  return { chips, onPress: (kind) => (kind === "name" ? actions.name() : actions.open()) };
}
