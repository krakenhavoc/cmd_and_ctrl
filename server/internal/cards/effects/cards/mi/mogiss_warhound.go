package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogis's Warhound — Enchantment Creature — Dog {1}{R}, 2/2 (Journey into Nyx):
//
//	"Bestow {2}{R}
//	 This creature attacks each combat if able.
//	 Enchanted creature gets +2/+2 and attacks each combat if able."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb87917f-33fb-4024-9317-0992c2f8f55e",
		Name:         "Mogis's Warhound",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{2}{R}"),
		},
		Static: []game.StaticAbility{
			effects.AttacksEachCombatWhere(effects.SelfCreatureOrAttached),
			effects.PumpAttached(2, 2),
		},
	})
}
