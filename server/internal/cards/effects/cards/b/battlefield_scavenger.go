package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefield Scavenger — Creature — Jackal Rogue {1}{R}, 2/2:
//
//	"You may exert this creature as it attacks. (It won't untap during
//	 your next untap step.)
//	 Whenever you exert a creature, you may discard a card. If you do,
//	 draw a card."
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with no linked
// trigger, and a "whenever you exert a creature" payoff that sees this
// creature and any other creature its controller exerts (the Amonkhet
// ruling). "You may discard a card. If you do, draw a card" is one
// instruction with a conditional second half (Cool but Rude's shape):
// the discard comes first and the draw follows only from it.
//
// No simplification.
func init() {
	const label = "Battlefield Scavenger — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:      "f88b0634-144a-4a33-a369-62df8edf05fe",
		Name:          "Battlefield Scavenger",
		Completeness:  effects.CompletenessFull,
		ExertOnAttack: effects.ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouExert(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label,
					effects.DrawAsManyAsDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
