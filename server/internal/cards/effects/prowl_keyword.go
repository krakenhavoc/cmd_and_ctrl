package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prowl_keyword.go — prowl (CR 702.76a, #5509): "You may pay [cost] rather
// than pay this spell's mana cost if a player was dealt combat damage this
// turn by a source that, at the time it dealt that damage, was under your
// control and had any of this spell's creature types."
//
// Freerunning's gated alternative cost (freerunning.go), with the gate
// reading the SPELL: its creature types are the card's own as it sits in
// the zone it is cast from, through the zone layers
// (AlternativeCost.SpellCondition), so a type an effect gives it there
// counts too (the Enigma Thief ruling), and a spell with every creature type
// is satisfied by a creature of any creature type. The dealers are the turn
// tally's record of what dealt combat damage to which player, by every
// subtype the dealer had and under whose control, at the moment it dealt
// the damage (PlayersDealtCombatDamageThisTurnBySubtype), so a Rogue that
// has since died, left your control or lost the type still counts, and so
// does a player who has left the game (the rulings). Checked at announce
// and in the view, never again once the spell is on the stack.

// Prowl is "Prowl <cost>".
//
//	AlternativeCosts: []game.AlternativeCost{Prowl("{3}{U}")}, // Enigma Thief
func Prowl(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:            "prowl",
		Label:          "Prowl " + cost,
		ManaCost:       cost,
		SpellCondition: prowlConditionMet,
	}
}

// prowlConditionMet is prowl's gate: a creature type of the spell among
// the creature types of a creature the caster controlled as it dealt
// combat damage to a player this turn.
func prowlConditionMet(g *game.Game, controller uuid.UUID, spell *game.Card) bool {
	for _, t := range game.CreatureTypesOf(spell) {
		if len(g.PlayersDealtCombatDamageThisTurnBySubtype(controller, t)) > 0 {
			return true
		}
	}
	return false
}
