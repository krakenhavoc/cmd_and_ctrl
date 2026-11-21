package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witty Roastmaster — Creature — Devil Citizen {2}{R}, 3/2 (EDHREC
// rank 1471):
//
//	"Alliance — Whenever another creature you control enters, this
//	 creature deals 1 damage to each opponent."
//
// Impact Tremors on a body. One trigger per creature entering, and
// the damage is the Roastmaster's, so a doubler doubles it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34edbf1b-0826-49f1-a04d-15cc8c543b22",
		Name:         "Witty Roastmaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CreatureEnterPaying(effects.WheneverAnotherCreatureEntersUnderYourControl("Witty Roastmaster — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			}), game.CreatureEnterPayoff{DamageEachOpponent: 1}),
		},
	})
}
