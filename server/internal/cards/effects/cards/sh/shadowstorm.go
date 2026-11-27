package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadowstorm — Sorcery {R} (Tempest, S82 OLD batch 757, #5628):
//
//	"Shadowstorm deals 2 damage to each creature with shadow."
//
// Squall's sweep with shadow in place of flying; keywords are read through
// the layered view, and the damage is one simultaneous instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e66e983-6349-428b-b511-8738c03ef8fc",
		Name:         "Shadowstorm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("shadow")), 2)
		},
	})
}
