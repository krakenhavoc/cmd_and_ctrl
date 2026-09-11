/**
 * faces.ts — the client half of the multi-face card model (ADR 0034).
 *
 * The wire keeps `name`, `type_line`, `mana_cost`, `power` and
 * `toughness` on CardView meaning "the ACTIVE face's", which is what
 * let twenty-odd client type checks keep working unchanged. What the
 * client needs on top of that is the handful of questions the face
 * PICKER asks: does this card need one, and what does it look like
 * once a different face is chosen?
 */

import type { CardView } from "./protocol";

/** Scryfall's layout for a modal double-faced card. */
export const LAYOUT_MODAL_DFC = "modal_dfc";

/**
 * needsFacePicker reports whether playing this card from hand
 * requires asking which half first.
 *
 * Only a modal DFC does (CR 712.12a — the faces are independently
 * playable). A transform card is always cast as its front face
 * (CR 712.4) and its back is reached by transforming the permanent;
 * adventure and split carry two faces on the wire but their second
 * halves are deferred. All three therefore report false, and so does
 * every single-faced card, which is why this gate opening the modal
 * costs nothing for the ~33,000 ordinary oracle IDs.
 *
 * Mirrors game.Card.CastableFaces server-side. If the two ever
 * disagree, the server wins: it refuses a face it did not offer with
 * ErrInvalidFace rather than silently casting the wrong half.
 */
export function needsFacePicker(card: CardView): boolean {
  return card.layout === LAYOUT_MODAL_DFC && (card.faces?.length ?? 0) > 1;
}

/**
 * cardAsFace returns a view of `card` as though face `i` were the
 * one up: the printed fields swapped for that face's, and
 * `active_face` moved so cardImageURL and every downstream type
 * check follow.
 *
 * The announce-prompt fields are CLEARED rather than carried over,
 * and that is the important half. `modes`, `additional_cost`,
 * `alternative_costs`, `tap_cost`, `legal_targets` and `target_mode`
 * are all computed server-side from the catalog spec of the face
 * that was active when the view was built — face 0. They describe
 * the FRONT half's rules and would be actively wrong attached to the
 * back. Dropping them means a back face currently announces with no
 * prompts, which is correct for all 60 land backs (a land has no
 * announce decisions at all) and honest for the 40 spell backs,
 * whose specs are not written yet. When they are, the server will
 * need to publish per-face prompt data and this is the function that
 * will consume it.
 */
export function cardAsFace(card: CardView, i: number): CardView {
  const face = card.faces?.[i];
  if (!face) return card;
  return {
    ...card,
    name: face.name,
    type_line: face.type_line,
    mana_cost: face.mana_cost,
    power: face.power,
    toughness: face.toughness,
    active_face: i,
    modes: undefined,
    additional_cost: undefined,
    alternative_costs: undefined,
    tap_cost: undefined,
    legal_targets: undefined,
    target_mode: undefined,
  };
}
