package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Putrefax — Creature — Phyrexian Horror {3}{G}{G}, 5/3 (Scars of
// Mirrodin, #5577):
//
//	"Trample, haste
//	 Infect
//	 At the beginning of the end step, sacrifice this creature."
//
// Three engine keywords (infect is the damage tail, CR 702.90) and
// Blistering Firecat's sacrifice at every player's end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3a336b30-335a-49d0-b192-c20f3670452b",
		Name:            "Putrefax",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste", "infect"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Putrefax — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
