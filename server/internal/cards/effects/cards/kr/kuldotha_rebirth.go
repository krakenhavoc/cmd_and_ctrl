package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kuldotha Rebirth — Sorcery {R}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Create three 1/1 red Goblin creature tokens."
//
// Hordeling Outburst's three Goblins behind Deadly Dispute's sacrifice cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f3e28778-0149-4dbd-badb-ba5aeacacd76",
		Name:           "Kuldotha Rebirth",
		Purpose:        game.Purpose{Tokens: 3, CreatureTokens: 3},
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact", effects.Artifact()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.RedGoblinToken(), N: 3}.Apply(ctx)
		},
	})
}
