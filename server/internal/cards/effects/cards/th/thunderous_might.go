package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thunderous Might — {1}{R} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, it gets +X/+0 until end of turn, where X is
//	 your devotion to red. (Each {R} in the mana costs of permanents you control counts
//	 toward your devotion to red.)"
//
// X is counted as the trigger resolves, from the Aura's controller's permanents
// (CR 700.5), as Theros' gods count it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d08117ec-e6ec-4c63-91d4-1bda59446c54",
		Name:         "Thunderous Might",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Thunderous Might — it gets +X/+0 until end of turn, where X is your devotion to red",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, ok := effects.EnchantedCreatureAsItResolves(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: host, Power: effects.DevotionToColors(g, item.Controller, "R"),
						Label: "Thunderous Might — +X/+0"}.Apply(ctx)
				}),
		},
	})
}
