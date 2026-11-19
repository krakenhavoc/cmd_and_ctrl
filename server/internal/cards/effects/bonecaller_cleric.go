package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonecaller Cleric — Creature — Human Cleric {1}{B}, 2/1 (Foundations
// Jumpstart):
//
//	"{3}{B}, Sacrifice this creature: Return target creature card from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe8d2874-c626-4d4e-a7f5-11c175a2b3bb",
		Name:         "Bonecaller Cleric",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}, Sacrifice this creature: Return target creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{B}"), SacrificeThis()),
			Targets:      targetCreatureInYourGraveyard(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				_, _ = reanimateSingleTarget(ctx, item.Controller)
				return nil
			},
		}},
	})
}
