package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prodigious Growth — Enchantment — Aura {4}{G}{G} (Core Set 2019, #5527):
//
//	"Enchant creature
//	 Enchanted creature gets +7/+7 and has trample."
//
// Frantic Strength's statics (PumpAttached, GrantToAttached) with +7/+7.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15d63e5d-dc4d-4e5f-8043-78039a49732b",
		Name:         "Prodigious Growth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(7, 7),
			effects.GrantToAttached("trample"),
		},
	})
}
