package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mighty Leap — Instant {1}{W} (Ravnica: Clue Edition, #5929):
//
//	"Target creature gets +2/+2 and gains flying until end of turn."
//
// Dive Down's body: the boost and the keyword land on the still-legal
// target (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "593caa15-e1d1-477f-bbb4-124d7a7b81f3",
		Name:         "Mighty Leap",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "flying")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(2, 2, "Mighty Leap — +2/+2 and flying", "flying")(ctx.Game, item)
		},
	})
}
