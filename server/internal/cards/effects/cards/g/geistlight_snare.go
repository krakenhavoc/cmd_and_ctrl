package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geistlight Snare — Instant {2}{U} (Innistrad: Crimson Vow, S82):
//
//	"This spell costs {1} less to cast if you control a Spirit. It also
//	 costs {1} less to cast if you control an enchantment.
//	 Counter target spell unless its controller pays {3}."
//
// Two separate self cost modifiers, so they stack (a Spirit enchantment
// still gives two {1}, as printed; CR 601.2f). The counter is Complicate's
// counterTheTargetSpellUnlessPaid: the spell's controller is asked as it
// resolves (CR 118.12a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45f159dc-a5f9-4335-896b-ad544c303801",
		Name:         "Geistlight Snare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Spirit.", effects.YouControlA(effects.OfCreatureType("Spirit"))),
			effects.CostsLess(1, "This spell costs {1} less to cast if you control an enchantment.", effects.YouControlA(effects.Enchantment())),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CounterTheTargetSpellUnlessPaid("Geistlight Snare", "{3}")(ctx.Game, item)
		},
	})
}
