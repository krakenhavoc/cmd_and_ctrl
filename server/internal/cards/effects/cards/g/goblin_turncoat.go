package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Turncoat — Creature — Goblin Mercenary {1}{B}, 2/1 (Dominaria
// Remastered):
//
//	"Sacrifice a Goblin: Regenerate this creature."
//
// Corrupted Harvester's regeneration with a sacrifice-a-Goblin cost. The
// Goblin sacrificed may be any Goblin its controller controls, the Turncoat
// itself included (it is a Goblin); regenerating it then does nothing, as
// it is already gone. Goblin is read through the layered subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "297c2ddf-6d2d-4ed4-ab37-716514d5e3b6",
		Name:         "Goblin Turncoat",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Goblin: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Goblin", effects.HasSubtype("Goblin"))},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
