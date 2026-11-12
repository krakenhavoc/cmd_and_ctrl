package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stoneskin — Enchantment — Aura {2}{W} (Commander Legends: Battle for
// Baldur's Gate, #5490):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +0/+10."
//
// Aquitect's Defenses' flash Aura without the enters trigger: a static
// +0/+10 on the host in layer 7c. It may enchant any creature, an
// opponent's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d1d53e3-60ee-40a1-89a5-0648760cd83f",
		Name:            "Stoneskin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         EnchantCreature(),
		Static:          []game.StaticAbility{PumpAttached(0, 10)},
	})
}
