package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "938e738f-c982-4bbf-b15f-0c3dd3a9901e",
		Name:         "Conclave's Blessing",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttachedPer(0, 2, vampirismOtherCreaturesYouControl),
		},
	})
}
