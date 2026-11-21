package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leaden Fists — Enchantment — Aura {2}{U} (Dominaria Remastered, #5948):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +3/+3 and doesn't untap during its controller's
//	 untap step."
//
// Flash is a printed keyword; the pump is PumpAttached and the untap clause
// is Bitter Chill's untap-step restriction on the enchanted creature (an
// untap effect still untaps it: only the untap step is skipped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d4bcbce9-9739-4828-9980-e21490b44346",
		Name:                  "Leaden Fists",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraBenefit},
		Static:                []game.StaticAbility{effects.PumpAttached(3, 3)},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
