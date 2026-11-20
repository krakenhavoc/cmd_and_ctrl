package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "5aa66cb0-86b9-4e85-9a17-df78416d682d",
		Name:         "Restrain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ShieldAgainstTheTargetThen(ctx, true, func(ctx *effects.Context, _ uuid.UUID) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			})
		},
	})
}
