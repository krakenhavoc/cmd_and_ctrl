package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archway Angel — Creature — Angel {5}{W}, 3/4 (Foundations, #4465):
//
//	"Flying
//	 When this creature enters, you gain 2 life for each Gate you
//	 control."
//
// Aven Gagglemaster's count with Gates in place of fliers: the Gates
// you control are counted as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f2cf087a-d89a-4145-9394-945ce41c8cca",
		Name:            "Archway Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Archway Angel — you gain 2 life for each Gate you control", func(g *game.Game, item *game.StackItem) error {
				gates := effects.PermanentsControlledByMatching(g, item.Controller, effects.HasSubtype("Gate"))
				return effects.GainLife{Amount: 2 * len(gates)}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
