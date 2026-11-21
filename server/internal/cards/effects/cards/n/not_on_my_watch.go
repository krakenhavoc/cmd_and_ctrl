package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Not on My Watch — Instant {1}{W} (Murders at Karlov Manor, #4469):
//
//	"Exile target attacking creature."
//
// Gaze of Justice's exile with the target narrowed to an attacking
// creature. A creature removed from combat in response is no longer a
// legal target, so the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d400c01-d2d0-442a-bc3d-cf106a754dab",
		Name:         "Not on My Watch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
