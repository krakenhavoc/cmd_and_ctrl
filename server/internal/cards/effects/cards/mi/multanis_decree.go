package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b233473d-d22b-45e4-9fe7-55d1170a788b",
		Name:         "Multani's Decree",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Enchantment(), ThenTail: ctx.Tail(effects.MassGainTwoLifeForEach)}.Apply(ctx)
		},
	})
}
