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

import type { CardView, CastSurfaceView } from "./protocol";

/** Scryfall's layout for a modal double-faced card. */
export const LAYOUT_MODAL_DFC = "modal_dfc";

/** Scryfall's layout for a creature with an Adventure (CR 715). */
export const LAYOUT_ADVENTURE = "adventure";

/**
 * needsFacePicker reports whether casting this card requires asking
 * which half first.
 *
 * Two layouts do. A modal DFC, because its faces are independently
 * playable (CR 712.12a) — the reason the picker exists. And an
 * adventure card, because CR 715.3 lets the caster choose between the
 * creature and the Adventure, which is the same question with a
 * different rules number.
 *
 * A transform card is always cast as its front face (CR 712.4) and
 * its back is reached by transforming the permanent; split carries two
 * faces on the wire but fusing is deferred. Both report false, and so
 * does every single-faced card, which is why this gate opening the
 * modal costs nothing for the ~33,000 ordinary oracle IDs.
 *
 * It does NOT take the source zone, and deliberately: which half a
 * GRANTED cast opens belongs to the grant, not to the card. The exile
 * button hands the settled face down with the cast (grantedFaceIndex),
 * and the picker is only reached when nothing has settled it — which
 * is an adventure card impulse-exiled by Ragavan, where both halves
 * really are open.
 *
 * Mirrors game.Card.CastableFaces server-side. If the two ever
 * disagree, the server wins: it refuses a face it did not offer with
 * ErrInvalidFace rather than silently casting the wrong half.
 */
export function needsFacePicker(card: CardView): boolean {
  if ((card.faces?.length ?? 0) < 2) return false;
  return card.layout === LAYOUT_MODAL_DFC || card.layout === LAYOUT_ADVENTURE;
}

/**
 * castSurfaceOf lifts one object's announce block out as a complete
 * key set — every key PRESENT, `undefined` included — so spreading it
 * over a CardView REPLACES that card's block rather than merging into
 * it.
 *
 * The distinction is the whole bug. `{...card, ...face}` would only
 * overwrite the keys the face happens to carry, so a face that
 * announces nothing would leave the front's `legal_targets` in place
 * — which is worse than clearing them, because the picker would open
 * on the wrong half's targets instead of not opening at all.
 *
 * `satisfies Record<keyof CastSurfaceView, unknown>` is what keeps
 * the list exhaustive: a field added to CastSurfaceView and forgotten
 * here is a type error rather than a field that quietly keeps the
 * front face's value. It is `satisfies` rather than a mapped
 * `{[K]-?: …}` annotation because `-?` strips `undefined` out of the
 * property types as well as the optionality, which is exactly the
 * value this has to be able to carry.
 */
export function castSurfaceOf(s: CastSurfaceView) {
  return {
    target_mode: s.target_mode,
    legal_targets: s.legal_targets,
    clauses: s.clauses,
    modes: s.modes,
    additional_cost: s.additional_cost,
    alternative_costs: s.alternative_costs,
    alternative_cost_required: s.alternative_cost_required,
    optional_costs: s.optional_costs,
    cant_cast: s.cant_cast,
    tap_cost: s.tap_cost,
    target_cost_notes: s.target_cost_notes,
    phyrexian_symbols: s.phyrexian_symbols,
    castable_here: s.castable_here,
  } satisfies Record<keyof CastSurfaceView, unknown>;
}

/**
 * cardAsFace returns a view of `card` as though face `i` were the
 * one up: the printed fields swapped for that face's, `active_face`
 * moved so cardImageURL and every downstream type check follow, and
 * the announce block swapped for THAT FACE's (#992).
 *
 * The announce block is the important half, and until #992 it was
 * CLEARED here rather than swapped. `modes`, `additional_cost`,
 * `alternative_costs`, `tap_cost`, `legal_targets`, `target_mode` and
 * the rest were computed server-side from the catalog spec of the
 * face that was active when the view was built — face 0 — so they
 * describe the FRONT half's rules and are actively wrong attached to
 * the back. Dropping them was correct while the only non-zero
 * castable faces were the sixty MDFC lands (a land has no announce
 * decisions) and forty spell backs with no specs written.
 *
 * #719 ended that. An adventure card's Adventure half is a real
 * castable face with a real Spec and most printed ones TARGET —
 * Stomp, Petty Theft, Swift End — so casting one from this client
 * announced with no target picker at all, and the server refused it.
 * The server publishes the announce data per castable face now
 * (protocol.CardFaceView), and this is the function that consumes it,
 * exactly as this docblock used to promise.
 *
 * A face the card offers no cast of carries no block, and the swap
 * then clears — which is the old behaviour, kept for the case it was
 * always right for: a transform card's back, or a half a grant does
 * not open.
 *
 * `hand_abilities` is still cleared rather than swapped. Cycling is
 * an ability of the CARD IN HAND (CR 702.29a) rather than of a face
 * being cast, the server stamps it per card, and a face swap has
 * nothing face-specific to put in its place.
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
    ...castSurfaceOf(face),
    hand_abilities: undefined,
  };
}
