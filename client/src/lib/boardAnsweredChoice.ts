// boardAnsweredChoice.ts — which pending choices are answered by
// clicking the BOARD rather than in a modal.
//
// One list, read by both halves of the decision. Board.svelte hands a
// matching choice to the targeting flow (the banner and the clickable
// permanents); ChoicePromptModal must NOT open for it. They used to
// keep separate lists: the board knew `legend_rule` and
// `choose_protector`, the modal skipped only `pick_target` and
// `retarget`, so a legend-rule prompt opened BOTH — and the modal's
// fallback branch rendered it as "Pick 1 card from opponent's revealed
// hand. opponent will discard your pick", under a full-screen backdrop
// that covered the board the real answer had to be clicked on (#1623).
//
// A new kind that is answered on the board goes here and nowhere else.

import { get } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { GameView, PendingChoiceView } from "./protocol";

const BOARD_ANSWERED = new Set<string>([
  // S20: a triggered ability's target, picked by clicking the board.
  "pick_target",
  // S27: the legend rule (CR 704.5j) — pick the permanent to keep.
  "legend_rule",
  // S27: a battle's protector (CR 310.9) — pick the player.
  "choose_protector",
  // #1196: the CR 115.7 retarget, answered on the board too.
  "retarget",
]);

export function isBoardAnsweredChoice(kind: string): boolean {
  return BOARD_ANSWERED.has(kind);
}

// #2394: the card-set picks that are answered on the board WHEN every
// candidate is a permanent on the battlefield — "untap up to five
// lands" (Finale of Revelation, a choose_cards over tapped lands) and
// the untap step's own "choose which of these untap" (untap_choice).
// The modal drew those lands as a grid of small rotated cards, and a
// player could not tell which was which. On the board they are the
// lands themselves, highlighted where they sit.
//
// Nothing about the question changes: the same candidates, the same
// choose_min / choose_max, and the same {choice_id, card_ids} answer
// (see choiceAnswer in targeting.ts). Only where it is asked.
const BOARD_PICKED_CARD_SETS = new Set<string>(["choose_cards", "untap_choice"]);

export function isBoardPickedCardSetKind(kind: string): boolean {
  return BOARD_PICKED_CARD_SETS.has(kind);
}

// listFallback holds the choices the player asked to see as the list
// instead ("Show as list" in the dock). Per client, never persisted: a
// choice ID is never reused, so the set only ever grows within a table.
export const listFallback = guardedWritable<ReadonlySet<string>>(new Set(), "listFallback");

export function showChoiceAsList(choiceID: string): void {
  listFallback.update((s) => new Set([...s, choiceID]));
}

// answeredOnBoard is the one predicate both halves read. `fallback` is
// the list-fallback set; it defaults to the store's current value.
export function answeredOnBoard(
  choice: PendingChoiceView,
  view: Pick<GameView, "battlefield">,
  fallback: ReadonlySet<string> = get(listFallback),
): boolean {
  if (isBoardAnsweredChoice(choice.kind)) return true;
  if (!isBoardPickedCardSetKind(choice.kind)) return false;
  if (fallback.has(choice.id)) return false;
  const options = choice.options ?? [];
  if (options.length === 0) return false;
  const onBattlefield = new Set(view.battlefield.cards.map((c) => c.instance_id));
  return options.every((o) => onBattlefield.has(o.instance_id));
}
