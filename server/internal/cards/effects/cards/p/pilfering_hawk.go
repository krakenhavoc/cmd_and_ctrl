package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pilfering Hawk — Snow Creature — Bird {1}{U}, 1/2 (Jumpstart 2022, #5948):
//
//	"Flying
//	 {S}, {T}: Draw a card, then discard a card. ({S} can be paid with one
//	 mana from a snow source.)"
//
// The loot is the shared draw-then-discard body (the discard is chosen after
// the draw); {S} is one mana from a snow source (CR 107.4h), as on Boreal
// Griffin.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8dbe7a06-3935-44a3-ac47-4ef00ed990f1",
		Name:            "Pilfering Hawk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{S}"), effects.TapCost()),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) },
		}},
	})
}
