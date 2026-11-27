package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niveous Wisps — Instant {W} (Shadowmoor, #5591):
//
//	"Target creature becomes white until end of turn. Tap that creature.
//	 Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4ecd56c-216d-4c3a-9dfa-09ae1be03e4f",
		Name:         "Niveous Wisps",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.OldB537WispsResolve("Niveous Wisps", "W", "white",
			func(ctx *effects.Context, t game.TargetRef) error {
				return effects.TapTarget{Target: t.ID}.Apply(ctx)
			}),
	})
}
