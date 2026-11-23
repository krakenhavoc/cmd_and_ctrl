package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spark Elemental — Creature — Elemental {R}, 3/1 (Fifth Dawn, #5604):
//
//	"Trample, haste (This creature can deal excess combat damage to the
//	 player or planeswalker it's attacking. This creature can attack
//	 and {T} as soon as it comes under your control.)
//	 At the beginning of the end step, sacrifice this creature."
//
// Hellspark Elemental's end-step sacrifice, which triggers at the beginning
// of every end step, not only its controller's, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "50608b7e-b6b2-40ca-91ec-db3e03dbb1af",
		Name:            "Spark Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Spark Elemental — sacrifice it", func(g *game.Game, item *game.StackItem) error {
				return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
