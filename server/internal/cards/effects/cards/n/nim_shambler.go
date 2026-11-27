package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nim Shambler — Creature — Zombie {2}{B}{B}, 2/1 (Mirrodin, #5606):
//
//	"This creature gets +1/+0 for each artifact you control.
//	 Sacrifice a creature: Regenerate this creature."
//
// Nim Devourer's artifact count, and Birthing Hulk's regeneration row with a
// sacrifice cost: any creature may be given up, the Shambler included.
//
// No simplification.
func init() {
	pump := effects.PumpSelfCreatureOrAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
		return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() })
	})
	pump.Label = "This creature gets +1/+0 for each artifact you control."
	effects.Register(effects.Spec{
		OracleID:     "83d2ff7f-4d6e-4cc0-b8a9-3c5b3236eea0",
		Name:         "Nim Shambler",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.SacrificeACreature(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
