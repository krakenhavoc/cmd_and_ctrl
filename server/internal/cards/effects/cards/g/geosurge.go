package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geosurge — Sorcery {R}{R}{R}{R} (New Phyrexia, S82 OLD batch 463, #5575):
//
//	"Add {R}{R}{R}{R}{R}{R}{R}. Spend this mana only to cast artifact or
//	 creature spells."
//
// Add-mana with the cast-only, artifact-or-creature spend restriction
// (Boseiju's "instant or sorcery" tag shape), seven red.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ec2794f-db1a-4448-a75e-ed7aff565378",
		Name:         "Geosurge",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.AddMana{
				Produced:     "{R}{R}{R}{R}{R}{R}{R}",
				Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictAnyType("Artifact", "Creature")},
			}.Apply(ctx)
		},
	})
}
