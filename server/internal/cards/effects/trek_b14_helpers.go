package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b14_helpers.go — shared pieces for Star Trek batch 14 (#4973).

// trekB14AltCostKeyBeamMeUp is beam me up's offer key: on the wire as
// `alternative_cost`, on StackItem.AltCost and on a permanent's
// CastProvenance. An on-disk identity: never renamed, never reused. The
// same string the other Star Trek batches' beam me up offers ride, so the
// keyword is one key however many files spell its constructor.
const trekB14AltCostKeyBeamMeUp = "beam_me_up"

// trekB14BeamMeUp is "Beam me up <cost> (You may cast this card from your
// graveyard for <cost> if you also return a creature you control to its
// owner's hand. Then exile this spell.)" (Boarding Action).
//
// The pinned Comprehensive Rules have no entry for the keyword, so the
// reminder text governs: flashback's cast from the graveyard and its
// exile as the spell leaves the stack (CR 702.34a's shape), priced at the
// mana plus Daze's return-a-permanent component narrowed to a creature
// you control. The return is a cost, matched and never targeted, paid
// with the spell already on the stack (CR 601.2h), so it stays paid if
// the spell is countered; a caster who controls no creature cannot claim
// the offer. The card file opens the graveyard:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{trekB14BeamMeUp("{2}{R}")},
func trekB14BeamMeUp(cost string) game.AlternativeCost {
	ac := Flashback(cost)
	ac.Key = trekB14AltCostKeyBeamMeUp
	ac.Label = "Beam me up " + cost
	ac.ReturnToHand = PermanentYouControl("a creature you control", Creature())
	ac.PayLabel = "a creature you control"
	return ac
}

// trekB14OfficerTokensYouControl is Academy Instructor's count, "each
// Officer token you control": the tokens the source's controller controls
// with the Officer subtype, read after the layers so a token that became
// an Officer counts and one that stopped being one does not.
func trekB14OfficerTokensYouControl(g *game.Game, source *game.Card) int {
	if source == nil {
		return 0
	}
	return len(permanentsControlledByMatching(g, source.Controller, And(Subtype("Officer"), IsTokenPredicate())))
}
