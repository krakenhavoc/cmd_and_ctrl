package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conclave's Blessing — Enchantment — Aura {3}{W} (Ravnica: City of
// Guilds, #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Enchant creature
//	 Enchanted creature gets +0/+2 for each other creature you control."
//
// Bravado's counted pump at +0/+2 with Lofty Dreams' convoke (CR 702.51a).
// "You" is the Aura's controller and "other" leaves out the enchanted
// creature, so the count is recomputed as creatures come and go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "938e738f-c982-4bbf-b15f-0c3dd3a9901e",
		Name:         "Conclave's Blessing",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(0, 2, effects.VampirismOtherCreaturesYouControl),
		},
	})
}
