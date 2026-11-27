package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silvergill Douser — Creature — Merfolk Wizard {1}{U}, 1/1 (Lorwyn, S82
// OLD b547, #5593):
//
//	"{T}: Target creature gets -X/-0 until end of turn, where X is the
//	 number of Merfolk and/or Faeries you control."
//
// X counts the permanents you control that are a Merfolk, a Faerie or both,
// each once (a changeling is both and counts once), as the ability resolves
// and fixed then (CR 611.2c). The Douser itself counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b24f2917-b96c-4348-b033-4926df3bd4b0",
		Name:         "Silvergill Douser",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets -X/-0 until end of turn, where X is the number of Merfolk and/or Faeries you control.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.S82OldB547CountYouControl(ctx, "Merfolk", "Faerie")
				return effects.BoostTheTargetUntilEOT(-x, 0, "Silvergill Douser — -X/-0")(item, ctx)
			},
		}},
	})
}
