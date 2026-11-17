package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ae19dce5-c9fc-4fbe-99b5-dc6a12124a68",
		Name:         "Akroma's Vengeance",
		Completeness: CompletenessFull,
		// ADR 0126 §6: artifacts, creatures and enchantments, the smallest listed class that holds all three.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Or(Artifact(), Creature(), Enchantment())}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
