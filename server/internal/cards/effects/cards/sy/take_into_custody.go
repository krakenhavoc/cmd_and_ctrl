package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Take into Custody — Instant {U} (Aether Revolt,
// #5536):
//
//	"Tap target creature. It doesn't untap during its controller's next
//	 untap step."
//
// Ojutai's Breath without rebound (TapAndFreeze).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e5f92b3-204b-42a3-b86f-8e2a1d263d1d",
		Name:         "Take into Custody",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Take into Custody")
		},
	})
}
