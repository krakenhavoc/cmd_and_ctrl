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
 * Scryfall's layout for a split card (CR 709) — Fire // Ice, the
 * aftermath and fuse cards, and every Room (ADR 0103).
 */
export const LAYOUT_SPLIT = "split";

/**
 * isAftermathFace reports whether face `i` of a split card is its
 * AFTERMATH half (CR 702.127a) — read, as the server reads it, off the
 * face's own oracle text, because the keyword is not a canonical
 * engine keyword.
 */
export function isAftermathFace(card: CardView, i: number): boolean {
  if (card.layout !== LAYOUT_SPLIT) return false;
  return /(^|\n)Aftermath\b/.test(card.faces?.[i]?.oracle_text ?? "");
}

/**
 * faceCastableFrom mirrors the server's split rule for which zones a
 * face may be cast out of (CR 702.127a): an aftermath half from a
 * graveyard and nowhere else. `zone` is a cast zone as the Board names
 * it — undefined is the hand. Every other face answers true; whether
 * the ZONE is open at all is the card's surfaces' business.
 */
export function faceCastableFrom(card: CardView, i: number, zone?: string): boolean {
  if (!isAftermathFace(card, i)) return true;
  return zone === "graveyard";
}

/**
 * displayName is a card's name for a label, and for every card but one
 * it is `card.name`. The one is a Room with both doors locked, which
 * has NO name (CR 709.5, ADR 0103 owner decision 6): the server sends
 * an empty name and the client labels it from its halves.
 */
export function displayName(card: CardView): string {
  if (card.name) return card.name;
  const faces = card.faces ?? [];
  if (card.layout === LAYOUT_SPLIT && faces.length >= 2) {
    return faces.map((f) => f.name).join(" // ");
  }
  return card.name ?? "";
}

/**
 * needsFacePicker reports whether casting this card requires asking
 * which half first.
 *
 * Two layouts do. A modal DFC, because its faces are independently
 * playable (CR 712.11b) — the reason the picker exists. And an
 * adventure card, because CR 715.3 lets the caster choose between the
 * creature and the Adventure, which is the same question with a
 * different rules number.
 *
 * A transform card is always cast as its front face (CR 712.11) and
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
  // ADR 0103: a split card's halves are cast one at a time (CR 709.3)
  // or, with fuse, together — a choice the picker asks too.
  return (
    card.layout === LAYOUT_MODAL_DFC ||
    card.layout === LAYOUT_ADVENTURE ||
    card.layout === LAYOUT_SPLIT
  );
}

/**
 * cardAsFused is a view of `card` as its FUSED cast (CR 702.102b, ADR
 * 0103): both names, both costs, and the server's announce block for
 * the fused cast — the left half's clauses then the right half's. The
 * card itself when it offers no fused cast.
 */
export function cardAsFused(card: CardView): CardView {
  const fused = card.fused;
  if (!fused) return card;
  return {
    ...card,
    name: fused.name,
    type_line: fused.type_line ?? card.type_line,
    mana_cost: fused.mana_cost,
    active_face: 0,
    ...castSurfaceOf(fused),
    zone_abilities: undefined,
    zone_mana_abilities: undefined,
  };
}

/**
 * FaceOption is one choice the face picker offers: a printed face, or
 * the fused cast of both halves of a split card with fuse.
 */
export interface FaceOption {
  face: number;
  fused: boolean;
  view: CardView;
}

/**
 * faceOptions is what the picker lists for a cast of `card` out of
 * `zone` (undefined is the hand): every face that zone allows
 * (faceCastableFrom), then the fused cast when the server published one.
 */
export function faceOptions(card: CardView, zone?: string): FaceOption[] {
  const out: FaceOption[] = [];
  (card.faces ?? []).forEach((_, i) => {
    if (faceCastableFrom(card, i, zone)) {
      out.push({ face: i, fused: false, view: cardAsFace(card, i) });
    }
  });
  if (card.fused && (zone === undefined || zone === "hand")) {
    out.push({ face: 0, fused: true, view: cardAsFused(card) });
  }
  return out;
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
    printed_cost_timing_closed: s.printed_cost_timing_closed,
    optional_costs: s.optional_costs,
    cant_cast: s.cant_cast,
    tap_cost: s.tap_cost,
    delve: s.delve,
    target_cost_notes: s.target_cost_notes,
    phyrexian_symbols: s.phyrexian_symbols,
    castable_here: s.castable_here,
    cast_prices: s.cast_prices,
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
 * `zone_abilities` is still cleared rather than swapped. Cycling is
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
    zone_abilities: undefined,
    // #1228: the mana half of the same clearing. A Spirit Guide has
    // one face, so nothing reaches this today; it is here because a
    // list left behind when the face swaps is the bug this function
    // exists to prevent.
    zone_mana_abilities: undefined,
  };
}

/**
 * castableFaces returns the views of `card` a cast may choose
 * between: `cardAsFace(card, i)` for every index `needsFacePicker`
 * would open a picker over, or `[card]` — the card exactly as handed
 * in — for every card that offers no such choice, single-faced,
 * transform and split alike.
 *
 * This is the ONE walk (#1173, #1168): both bugs were the same
 * mistake in two files, reading a multi-face card's own top-level
 * block for a question that is really "does ANY face answer yes".
 * That block is the face that HAPPENS TO BE UP — always face 0, for a
 * card in hand or a front-up graveyard pile (CR 712.8a) — and answering
 * from it alone silently drops every other face's answer. Sharing this
 * enumeration is what keeps `castableFromZone` (zoneBrowser.logic.ts)
 * and `canCastFromHand`'s announce gates (timing.ts) from drifting
 * into two different opinions about which faces a cast may choose.
 *
 * Collapses to `[card]` for the ~33,000 ordinary single-faced oracle
 * IDs, and for a transform or split card, exactly as `needsFacePicker`
 * does — cardAsFace(card, 0) is already the same answer `card` is
 * (see its own docblock), so returning `[card]` rather than
 * `[cardAsFace(card, 0)]` for those costs nothing.
 */
export function castableFaces(card: CardView, zone?: string): CardView[] {
  if (!needsFacePicker(card)) return [card];
  return faceOptions(card, zone).map((o) => o.view);
}

/**
 * castableFaceIndex picks which of `castableFaces(card)` a caller
 * should treat as the default — the face picker's initial highlight,
 * chief among callers — by running `isCastable` over them in printed
 * order and returning the first index it accepts.
 *
 * `isCastable` is supplied by the caller rather than fixed here
 * because "castable" means a different thing to each of the two bugs
 * this shares its enumeration with: a graveyard permission answers
 * from `castable_here`, a hand cast from whether the face's own
 * target clause has anything to point at. Neither reader belongs in
 * faces.ts — the graveyard one already lives beside `castable_here`'s
 * other reader in zoneBrowser.logic.ts, and the hand one beside
 * `hasSatisfiableTargets` in timing.ts — so this stays a leaf module
 * neither needs to import.
 *
 * Falls back to 0 — the server's own default for a cast with no
 * `face` at all — when no face satisfies `isCastable`, which includes
 * every single-faced card: `castableFaces` returns just `[card]` for
 * those, and `card` is exactly what `isCastable` is asked about.
 */
export function castableFaceIndex(
  card: CardView,
  isCastable: (face: CardView) => boolean,
  zone?: string,
): number {
  const i = castableFaces(card, zone).findIndex(isCastable);
  return i === -1 ? 0 : i;
}
