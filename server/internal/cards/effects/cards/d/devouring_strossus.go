package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devouring Strossus (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Phyrexian Horror {5}{B}{B}{B}, 9/9:
//	"Flying, trample
//	At the beginning of your upkeep, sacrifice a creature.
//	Sacrifice a creature: Regenerate this creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8f2ab0a8-9d60-493b-a284-9ed7fa07e5ea",
		Name:            "Devouring Strossus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Devouring Strossus — sacrifice a creature", effects.S82OldB646SacrificeACreature("Devouring Strossus")),
		},
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
