package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heroic Defiance — Enchantment — Aura {1}{W} (Planeshift, #5615):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 unless it shares a color with the most
//	 common color among all permanents or a color tied for most common."
//
// A layer-7c +3/+3 under a condition read afresh on every recompute (#6207's
// SharesAMostCommonColor over every permanent's layered colours, CR 105.2):
// the Aura itself counts among the permanents, a colourless creature shares no
// colour and so is pumped, and a creature that gains or loses a colour gains or
// loses the pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6209f18-9f40-44b4-9a8c-3445eda89928",
		Name:         "Heroic Defiance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.S82OldB636NotSharingAMostCommonColor, 3, 3),
		},
	})
}
