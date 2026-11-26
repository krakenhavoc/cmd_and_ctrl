package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giant Spectacle — Enchantment — Aura {1}{R} (Kaladesh, #5539):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has menace."
//
// Beamtown Beatstick's two statics on the enchanted creature: a layer 7c pump
// and a layer 6 menace grant, both for as long as the Aura is attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97e34c3a-070e-406a-8505-657f4d57def8",
		Name:         "Giant Spectacle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttached("menace"),
		},
	})
}
