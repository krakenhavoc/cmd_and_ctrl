package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horrid Vigor — Instant {1}{G} (Duskmourn: House of Horror, #4466):
//
//	"Target creature gains deathtouch and indestructible until end of
//	 turn."
//
// Give In to Violence's body with no pump: both keywords land on the
// still-legal target (CR 608.2b) until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bef3f0a9-4d5c-4469-81df-adbeedeed768",
		Name:         "Horrid Vigor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(0, 0, "Horrid Vigor — deathtouch and indestructible", "deathtouch", "indestructible")(ctx.Game, item)
		},
	})
}
