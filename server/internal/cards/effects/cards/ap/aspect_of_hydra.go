package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aspect of Hydra — Instant {G} (Born of the Gods, S82 OLD b377, #5559):
//
//	"Target creature gets +X/+X until end of turn, where X is your devotion to
//	 green."
//
// Devotion is counted once as the spell resolves (CR 700.5), through the
// shared S82OldB133DevotionBoostFirstTarget.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1580692-ca1c-4682-8cad-7ad380b39454",
		Name:         "Aspect of Hydra",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB133DevotionBoostFirstTarget("Aspect of Hydra — +X/+X", "G", 1, 1)(ctx.Game, item)
		},
	})
}
