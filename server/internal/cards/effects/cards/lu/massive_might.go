package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Massive Might — Instant {G} (Innistrad: Crimson Vow):
//
//	"Target creature gets +2/+2 and gains trample until end of turn."
//
// Dive Down's shape: the boost and the keyword land on the still-legal
// target (CR 608.2b) and end in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f0ec90b-e5a0-49bd-8fa0-9be6ba259a62",
		Name:         "Massive Might",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(2, 2, "Massive Might — +2/+2 and trample", "trample")(ctx.Game, item)
		},
	})
}
