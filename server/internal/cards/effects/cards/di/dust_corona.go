package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dust Corona — Enchantment — Aura {R} (Planar Chaos, #5595):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0 and can't be blocked by creatures with
//	 flying."
//
// Fire-Rim Form's +2/+0 (PumpAttached) and Skyscythe Engulfer's block
// restriction moved to the enchanted creature (CantBeBlockedBy over
// OnAttached): both follow the Aura and end the moment it leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f9ab49b-81fc-487e-8613-e30afcae164d",
		Name:         "Dust Corona",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnAttached(), effects.HasKeyword("flying"), "creatures with flying"),
		},
	})
}
