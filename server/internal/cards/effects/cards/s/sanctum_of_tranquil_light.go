package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum of Tranquil Light — Legendary Enchantment — Shrine {W} (Core Set
// 2021, #5511):
//
//	"{5}{W}: Tap target creature. This ability costs {1} less to activate for
//	 each Shrine you control."
//
// Battlefield Butcher's ability cost reduction (CR 602.2b, 601.2f) counted
// over the Shrines you control; the Sanctum is itself a Shrine, so the
// ability costs at least {1} less. Only the generic mana is reduced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1818b0b1-0104-4673-9be8-acfc773ef87d",
		Name:         "Sanctum of Tranquil Light",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{W}: Tap target creature. This ability costs {1} less to activate for each Shrine you control.",
			Cost:    effects.ManaCost("{5}{W}"),
			Targets: effects.TargetCreature("target creature"),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(func(q game.CostQuery) int {
					if q.Game == nil {
						return 0
					}
					return effects.CountControlled(q.Game, q.Controller, func(c game.Card) bool { return c.HasSubtype("Shrine") })
				}, "This ability costs {1} less to activate for each Shrine you control."),
			},
			Effect: effects.TapChosenPermanent,
		}},
	})
}
