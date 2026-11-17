package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bravado — Enchantment — Aura {1}{R} (Urza's Saga):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each other creature you control."
//
// Vampirism's first static, word for word: a layer 7c pump counted on
// every recompute. "You" is the Aura's controller, and "other" leaves out
// the enchanted creature itself, so on an opponent's creature the count
// is every creature you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69b5af09-b3e6-4abf-b62e-9199c4f43ba7",
		Name:         "Bravado",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, vampirismOtherCreaturesYouControl),
		},
	})
}
