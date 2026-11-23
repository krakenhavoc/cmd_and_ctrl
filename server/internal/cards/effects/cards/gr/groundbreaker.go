package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Groundbreaker — Creature — Elemental {G}{G}{G}, 6/1 (Planar Chaos, #5595):
//
//	"Trample, haste
//	 At the beginning of the end step, sacrifice this creature."
//
// Hellspark Elemental's end-step sacrifice: "the end step" is every
// player's, so it goes at the end of the turn it was cast in, whoever's
// turn that is (AtEachStep).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a14dc46f-b790-458e-8821-0c5646188680",
		Name:            "Groundbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Groundbreaker — sacrifice it", func(g *game.Game, item *game.StackItem) error {
				return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
