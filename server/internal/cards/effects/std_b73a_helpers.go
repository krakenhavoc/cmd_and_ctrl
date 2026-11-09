package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b73a_helpers.go — shared pieces for Lorwyn Eclipsed batch 73a
// (#4458).

// YouCastASpellOfTheChosenType is "whenever you cast a spell of the
// chosen type" (Door of Destinies, Chronicle of Victory): the source's
// controller cast it, and the spell has the creature type chosen for the
// source (Card.NamedTribe) as it is on the stack, so a kindred spell or a
// changeling of the type counts. A source with no type chosen matches
// nothing. Not "a creature spell": Vanquisher's Banner's reads that.
func YouCastASpellOfTheChosenType(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller || source.NamedTribe == "" {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.HasSubtype(source.NamedTribe)
}
