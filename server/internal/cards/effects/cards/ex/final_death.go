package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final Death — Instant {4}{B} (Theros Beyond Death, #5514):
//
//	"Exile target creature."
//
// Wander Off's text: the target is exiled if it is still legal as the
// spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1be3eea4-eaff-469c-9689-abadac1e848c",
		Name:         "Final Death",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTheFirstTargetIfLegal(ctx)
		},
	})
}
