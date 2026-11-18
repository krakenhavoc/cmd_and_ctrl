package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Public Enemy — Enchantment — Aura {2}{U}:
//
//	"Enchant creature
//	 All creatures attack enchanted creature's controller each combat if
//	 able.
//	 When enchanted creature dies, draw a card."
//
// The requirement belongs to the Aura (a creature that loses its abilities
// still has to attack) and names the enchanted creature's CURRENT
// controller, read as each declaration is judged, so a stolen host moves
// the target with it. The dies trigger is the Aura's own: its controller
// draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5f360dc-fe4e-4796-b5c1-468cbd0f6d24",
		Name:         "Public Enemy",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{s82SNC22AllCreaturesAttackHostController()},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Public Enemy — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
