package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pledge of Loyalty — Enchantment — Aura {1}{W} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 Enchanted creature has protection from each color among permanents you
//	 control. This effect doesn't remove this Aura."
//
// Flickering Ward's sparing protection (CR 702.16n), one instance per color
// among the Aura controller's permanents, re-read on every layer pass. The
// Aura is itself white, so it makes the creature protected from white; each
// instance is minted with the Aura's id so it stays attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e88b3066-eb10-4ade-8207-af4de40facec",
		Name:         "Pledge of Loyalty",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.OldB661ProtectionFromEachColorAmongYourPermanentsSparingThisAura(),
		},
	})
}
