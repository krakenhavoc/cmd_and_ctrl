package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Offer Immortality — Instant {1}{B} (Foundations, #4465):
//
//	"Target creature gains deathtouch and indestructible until end of
//	 turn."
//
// The still-legal target gains both keywords until end of turn
// (jmpB7FirstLegalTargetGains, Duelist's Heritage's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be4499e6-616a-4c11-b37e-9201d2492682",
		Name:         "Offer Immortality",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.JmpB7FirstLegalTargetGains("Offer Immortality — deathtouch and indestructible",
				"deathtouch", "indestructible")(ctx.Game, item)
		},
	})
}
