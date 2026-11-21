package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inkrise Infiltrator — Creature — Human Ninja {1}{B}, 1/2 (Neon Dynasty):
//
//	"Flying
//	 {3}{B}: This creature gets +2/+2 until end of turn."
//
// Flying is the printed keyword. The pump is Armored Armadillo's shape
// with a fixed bonus: a source that left, or came back as a new object,
// gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a45d09a-6af3-45fb-8f7c-cf630cb37f24",
		Name:            "Inkrise Infiltrator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.ManaCost("{3}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.ThisGetsUntilEndOfTurn(2, 2, "Inkrise Infiltrator — +2/+2 until end of turn")(g, item)
			},
		}},
	})
}
