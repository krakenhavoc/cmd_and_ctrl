package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Madcap Skills — Enchantment — Aura {1}{R} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+0 and has menace."
//
// Swashbuckling's shape: a layer 7c pump and a layer 6 keyword grant that follow the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a9dc9ef-a603-467a-aeed-32cc111394b0",
		Name:         "Madcap Skills",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.GrantToAttached("menace"),
		},
	})
}
