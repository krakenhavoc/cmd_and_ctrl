package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surge of Righteousness — Instant {1}{W} (Dragons of Tarkir):
//
//	"Destroy target black or red creature that's attacking or blocking. You gain
//	 2 life."
//
// The life is gained only when the spell resolves with its target still legal
// (CR 608.2b): a target that went away or stopped being black or red, attacking
// or blocking fizzles the whole spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c9a2789-4b8c-4706-a95a-99561cd6d09c",
		Name:         "Surge of Righteousness",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 2},
		Targets: effects.TargetCreature("target black or red creature that's attacking or blocking",
			effects.And(effects.Or(effects.OfColor("B"), effects.OfColor("R")), effects.AttackingOrBlocking())),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			}),
			effects.GainLife{Amount: 2},
		),
	})
}
