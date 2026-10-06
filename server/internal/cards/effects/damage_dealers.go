package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damage_dealers.go — "creatures that dealt (combat) damage to you
// this turn" (#2149, Witch-king of Angmar, Reciprocate, Retaliate,
// Spear of Heliod).
//
// The record is TurnTally.DamageDealers: one entry per creature OBJECT
// per victim, written as the damage is dealt, so a creature that has
// since died or been flickered is answered correctly (CR 400.7: the
// flickered one is a new object that dealt nothing). The predicates
// read it for the player the predicate is handed as the caster — the
// controller of the spell or ability — which is "you" on every card
// printed with this clause. A clause whose predicate is not handed the
// controller (an edict, a mass effect) fixes the victim when the
// ability resolves and uses the two functions below.

// DealtDamageToYouThisTurn is the creature, as it is now, having dealt
// damage of any kind to the caster this turn: "target creature that
// dealt damage to you this turn". Only a creature's own damage counts;
// a burn spell or a noncreature permanent is not such a creature.
func DealtDamageToYouThisTurn() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return g.ObjectDealtDamageToPlayerThisTurn(c.InstanceID, caster)
	}
}

// DealtCombatDamageToYouThisTurn is the same for combat damage only:
// "a creature that dealt combat damage to you this turn".
func DealtCombatDamageToYouThisTurn() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return g.ObjectDealtCombatDamageToPlayerThisTurn(c.InstanceID, caster)
	}
}

// dealtDamageToPlayerThisTurn is DealtDamageToYouThisTurn for a victim
// fixed by the caller, for the mass effects whose predicate is not
// handed the controller.
func dealtDamageToPlayerThisTurn(victim uuid.UUID) CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.ObjectDealtDamageToPlayerThisTurn(c.InstanceID, victim)
	}
}

// dealtCombatDamageToPlayerThisTurn is its combat-only twin, for an
// edict: the sacrificing player is not the one the damage was dealt to.
func dealtCombatDamageToPlayerThisTurn(victim uuid.UUID) CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.ObjectDealtCombatDamageToPlayerThisTurn(c.InstanceID, victim)
	}
}
