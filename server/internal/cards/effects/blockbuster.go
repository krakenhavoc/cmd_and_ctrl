package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "11e50e35-bf4b-4089-b0c9-360d9c014584",
		Name:         "Blockbuster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{R}, Sacrifice this enchantment: It deals 3 damage to each tapped creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove,
				Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
			Cost: Plus(ManaCost("{1}{R}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageEachMatchingAndEachPlayerThen(NewContext(g, item), And(Creature(), Not(Untapped())), 3, RecipientThen{})
			},
		}},
	})
}
