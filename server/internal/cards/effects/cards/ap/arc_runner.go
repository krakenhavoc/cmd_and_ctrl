package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arc Runner — Creature — Elemental Ox {2}{R}, 5/1 (Magic 2011, #5579):
//
//	"Haste
//	 At the beginning of the end step, sacrifice this creature."
//
// Lightning Serpent's end-step sacrifice: the end step of every turn,
// not only its controller's, so it is gone at the end of the turn it
// arrived. Haste is the printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9bbee1a5-2eb0-4fad-9524-4ad9cc5ca998",
		Name:            "Arc Runner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Arc Runner — sacrifice it", func(g *game.Game, item *game.StackItem) error {
				return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
