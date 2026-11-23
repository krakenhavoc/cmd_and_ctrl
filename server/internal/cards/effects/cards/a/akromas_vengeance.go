package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akroma's Vengeance — Sorcery {4}{W}{W} (Onslaught):
//
//	"Destroy all artifacts, creatures, and enchantments.
//	 Cycling {3}"
//
// Nevinyrral's Disk's sweep as a sorcery: one destroy over the
// three-way Or, so an artifact creature is destroyed once. Lands and
// planeswalkers stay. Regeneration and indestructible apply as usual.
// Cycling is the keyword's own hand ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae19dce5-c9fc-4fbe-99b5-dc6a12124a68",
		Name:         "Akroma's Vengeance",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: artifacts, creatures and enchantments, the smallest listed class that holds all three.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
