package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonewood Invoker — Creature — Elf Mutant {1}{G}, 2/2 (Dominaria
// Remastered):
//
//	"{7}{G}: This creature gets +5/+5 until end of turn."
//
// A plain pump on the source, pinned to this object for the turn
// (BoostUntilEOT), so a copy that left and came back is not pumped
// (CR 400.7). Each activation stacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d277206-9156-4dc9-bdd2-9d9219787735",
		Name:         "Stonewood Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{G}: This creature gets +5/+5 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 5, Toughness: 5}},
			Cost:    effects.ManaCost("{7}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 5, Toughness: 5, Label: "Stonewood Invoker — +5/+5 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
