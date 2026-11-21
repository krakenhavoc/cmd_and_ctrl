package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystalline Armor — Enchantment — Aura {3}{G} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each land you control and has
//	 trample."
//
// Blackblade Reforged's count on an Aura: a layer 7c pump (CR 613.4c)
// of one per land the AURA's controller controls (CR 303.4e), read
// continuously, plus a trample grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93d2638b-0c19-4fba-a0ad-84b6a9784eb3",
		Name:         "Crystalline Armor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, effects.LandsControlledBy),
			effects.GrantToAttached("trample"),
		},
	})
}
