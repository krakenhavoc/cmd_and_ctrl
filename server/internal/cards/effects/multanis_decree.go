package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Multani's Decree — Sorcery {3}{G} (Urza's Destiny, #5621):
//
//	"Destroy all enchantments. You gain 2 life for each enchantment
//	 destroyed this way."
//
// Fumigate's shape over enchantments, with Righteous Fury's two-life
// tail: the life counts only the enchantments that were actually
// destroyed, so one that was regenerated or that a replacement kept
// adds nothing. Every player's enchantments go, the caster's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b233473d-d22b-45e4-9fe7-55d1170a788b",
		Name:         "Multani's Decree",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Enchantment(), ThenTail: ctx.Tail(massGainTwoLifeForEach)}.Apply(ctx)
		},
	})
}
