package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Restrain — Instant {2}{W}:
//
//	"Prevent all combat damage that would be dealt by target attacking creature this turn.
//	 Draw a card."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): Warning with a card. The draw
// happens whenever the spell resolves; with its only target gone it does
// not resolve at all (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5aa66cb0-86b9-4e85-9a17-df78416d682d",
		Name:         "Restrain",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return shieldAgainstTheTargetThen(ctx, true, func(ctx *Context, _ uuid.UUID) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			})
		},
	})
}
