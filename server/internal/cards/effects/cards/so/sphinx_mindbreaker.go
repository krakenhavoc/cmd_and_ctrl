package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx Mindbreaker — Creature — Sphinx {5}{U}{U}, 6/6 (Theros Beyond Death, #5514):
//
//	"Flying
//	 When this creature enters, each opponent mills ten cards."
//
// Flying is the printed keyword; the enters trigger is Fractured
// Sanity's each-opponent mill, ten cards each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a5270cf7-1eb3-4907-9f47-d787f61940d3",
		Name:            "Sphinx Mindbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sphinx Mindbreaker — each opponent mills ten cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B12EachOpponentMills(g, item, 10)
				}),
		},
	})
}
