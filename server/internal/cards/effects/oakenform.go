package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oakenform — Enchantment — Aura {2}{G} (Magic 2010, #5585):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3."
//
// Gift of Strands' pump without the scry: the Aura targets a creature as
// it is cast, and the bonus is a layer 7c static on the enchanted creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b254fefb-11d2-4369-ae5c-0c3bd48ae580",
		Name:         "Oakenform",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
	})
}
