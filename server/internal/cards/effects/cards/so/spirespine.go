package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirespine — Enchantment Creature — Beast {2}{G}, 4/1 (Journey into Nyx):
//
//	"Bestow {4}{G}
//	 This creature blocks each combat if able.
//	 Enchanted creature gets +4/+1 and blocks each combat if able."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de0edcc8-1595-4dbb-ac8a-46dcd96299cb",
		Name:         "Spirespine",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{4}{G}"),
		},
		Static: []game.StaticAbility{
			effects.BlocksEachCombatWhere(effects.SelfCreatureOrAttached),
			effects.PumpAttached(4, 1),
		},
	})
}
