package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b6_helpers.go — shared pieces for Star Trek batch 6 (#4973).

// trekB6BeamMeUpKey is the alternative-cost key "Beam me up" rides on
// StackItem.AltCost. An on-disk identity: never renamed, never reused.
const trekB6BeamMeUpKey = "beam_me_up"

// trekB6BeamMeUp is "Beam me up <cost> (You may cast this card from your
// graveyard for <cost> if you also return a creature you control to its
// owner's hand. Then exile this spell.)" (Star Trek). Flashback's cast
// from the graveyard and its exile on leaving the stack (CR 702.34a's
// shape), priced at the mana plus Daze's return-a-permanent component
// narrowed to a creature you control. The bounce is a cost, so it is paid
// with the spell on the stack and stays paid if the spell is countered.
// It is keyed apart from flashback so a card that asks about a flashback
// cast does not see it. The card file opens the graveyard:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{trekB6BeamMeUp("{3}{U}")},
func trekB6BeamMeUp(cost string) game.AlternativeCost {
	ac := Flashback(cost)
	ac.Key = trekB6BeamMeUpKey
	ac.Label = "Beam me up " + cost
	ac.ReturnToHand = PermanentYouControl("a creature you control", Creature())
	ac.PayLabel = "a creature you control"
	return ac
}

// ownerTucksFirstTargetTopOrBottom is "Target <permanent>'s owner puts it
// on their choice of the top or bottom of their library" (Run Behind,
// Guidance Failure): the first still-legal target goes to its owner's
// library, the owner choosing the end. A target gone by resolution does
// nothing.
func ownerTucksFirstTargetTopOrBottom(ctx *Context, label string) error {
	var id uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			id = t.ID
			break
		}
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.Owner == uuid.Nil {
		return nil
	}
	return PutInLibraryInAnyOrder{
		Chooser:   c.Owner,
		Cards:     []uuid.UUID{id},
		From:      game.ZoneBattlefield,
		Placement: game.LibraryPlaceTopOrBottom,
		Label:     label,
	}.Apply(ctx)
}
