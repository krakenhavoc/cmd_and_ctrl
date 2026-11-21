package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonecaller Cleric — Creature — Human Cleric {1}{B}, 2/1 (Foundations
// Jumpstart):
//
//	"{3}{B}, Sacrifice this creature: Return target creature card from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe8d2874-c626-4d4e-a7f5-11c175a2b3bb",
		Name:         "Bonecaller Cleric",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B}, Sacrifice this creature: Return target creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{B}"), effects.SacrificeThis()),
			Targets:      effects.TargetCreatureInYourGraveyard(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				_, _ = effects.ReanimateSingleTarget(ctx, item.Controller)
				return nil
			},
		}},
	})
}
