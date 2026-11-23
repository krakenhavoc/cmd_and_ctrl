package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blockbuster — Enchantment {3}{R}{R} (Ravnica: City of Guilds, #5600):
//
//	"{1}{R}, Sacrifice this enchantment: It deals 3 damage to each tapped
//	 creature and each player."
//
// Flamebreak's body with "tapped" in place of "without flying": the
// creatures tapped as the ability resolves, then every player still in the
// game, the controller included. Blockbuster is sacrificed as the cost, so
// it deals the damage as it last existed on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11e50e35-bf4b-4089-b0c9-360d9c014584",
		Name:         "Blockbuster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{R}, Sacrifice this enchantment: It deals 3 damage to each tapped creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove,
				Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
			Cost: effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatchingAndEachPlayerThen(effects.NewContext(g, item), effects.And(effects.Creature(), effects.Not(effects.Untapped())), 3, effects.RecipientThen{})
			},
		}},
	})
}
