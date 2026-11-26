package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paraselene — {2}{W} Sorcery (Innistrad, S82 OLD b447, #5572):
//
//	"Destroy all enchantments. You gain 1 life for each enchantment destroyed
//	 this way."
//
// Bane of Progress's keyed sweep: the life counts only the enchantments that
// were actually destroyed (an indestructible one stays and counts for nothing),
// told as the pre-move copies after any replacement prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e14ac0f3-00d5-4a95-990c-b7ea157bb87e",
		Name:         "Paraselene",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Enchantment(), ThenTail: ctx.ItemTail(paraseleneLife)}.Apply(ctx)
		},
	})
}

// paraseleneLife is "you gain 1 life for each enchantment destroyed this way".
// An on-disk key: never renamed, never reused.
var paraseleneLife = effects.MassTail("mass/paraselene-life", func(ctx *effects.Context, _ []game.Card, destroyed int) error {
	return effects.GainLife{Player: ctx.Controller(), Amount: destroyed}.Apply(ctx)
})
