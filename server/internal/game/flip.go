package game

import "github.com/google/uuid"

// flip.go — CR 710, flipping a flip card (ADR 0151).
//
// A flip card is ONE face printed the right way up and an alternative
// set of characteristics printed upside down (CR 710.1). The deck
// importer already gives it two game.Faces; this file is the verb that
// turns it over, and Card.Flipped is the CR 110.5 status the rules name.
//
// Flipping is not transforming. CR 701.27 is about double-faced cards
// and a flip card is single-faced, so TransformPermanentForEffect keeps
// refusing one (isDoubleFacedPermanent) and CanFlip refuses every
// double-faced card. Two verbs, two rules, and neither reaches the
// other.
//
// The shape is TransformPermanentForEffect's, for the same rule: a
// flip is not a zone change (CR 710.1c's "any changes to it by external
// effects will still apply" is CR 712.18 in other words). The permanent
// keeps its InstanceID, ObjectEpoch, counters, damage, attachments, tap
// state, controller, CR 613.7 timestamp and summoning sickness, and
// nothing re-triggers.

// CanFlip reports whether an instruction to flip this permanent does
// anything: a face-up flip card that is not already flipped (CR 710.4).
//
// Three refusals:
//
//   - a permanent already flipped: flipping is one-way (CR 710.4), and
//     flipping twice is nothing;
//   - a face-down permanent: it has no text (CR 708.2a), so nothing it
//     prints could have asked;
//   - a permanent that is not a flip card. CR 110.5c lets any permanent
//     carry the status, irrelevant to it, but every printed flip
//     instruction is a flip card's "flip this creature", and no card
//     flips another permanent (ADR 0151, out of scope 2).
func CanFlip(c Card) bool {
	if c.Flipped || c.FaceDownIsPermanent() {
		return false
	}
	return IsFlipCard(c)
}

// FlipPermanentForEffect flips the named battlefield permanent (CR 710):
// it becomes Flipped and, a flip card, shows its alternative half (CR
// 710.2) — the name, type line, P/T, keywords and catalog key
// ("<oracle_id>#1") change; its colour and mana cost do not (CR 710.1c,
// SetFace).
//
// Returns nil and emits nothing for a permanent CanFlip refuses: an
// instruction that legally does nothing is a resolution, not an
// EventEffectError (ADR 0079 decision 4). ErrCardNotFound for an ID
// naming no battlefield permanent, a caller holding a stale ID.
//
// The effective cache is nilled here and EventFlipPermanent bumps the
// layer version (layerVersionBump), for TransformPermanentForEffect's
// two reasons.
//
// Caller must hold g.mu.
func (g *Game) FlipPermanentForEffect(cardID uuid.UUID) error {
	card := findBattlefieldCard(g, cardID)
	if card == nil {
		return ErrCardNotFound
	}
	if !CanFlip(*card) {
		return nil
	}
	from := card.Name
	card.Flipped = true
	card.SetFace(1)
	card.effective = nil
	g.EmitEvent(Event{
		Kind:   EventFlipPermanent,
		CardID: cardID,
		Actor:  card.Controller,
		Label:  from,
	})
	return nil
}

// ReturnFromGraveyardFlippedForEffect is "return it to the battlefield
// flipped" (Homura, Human Ascendant), the one printed exception to CR
// 110.5b's "permanents enter unflipped". The transform form's body
// (returnFromGraveyardOnFaceLocked) with the flip card's guard: the
// card is set flipped, on its flipped half, while it is still in the
// graveyard, and both are put back if the entry is cancelled or
// redirected. A card that is not a flip card stays in the graveyard,
// since "flipped" has nothing to make of it and an unflipped return
// would be the stronger-than-printed outcome (#259).
//
// controller uuid.Nil means its owner's control. ErrCardNotFound when
// the card is not in a graveyard any more (CR 400.7). Returns the
// permanent's ID, or uuid.Nil when nothing entered (or it paused).
//
// Caller must hold g.mu.
func (g *Game) ReturnFromGraveyardFlippedForEffect(cardID, controller uuid.UUID, tapped bool) (uuid.UUID, error) {
	return g.returnFromGraveyardOnFaceLocked(cardID, 1, true, IsFlipCard, controller, tapped, nil)
}
