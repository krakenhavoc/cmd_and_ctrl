package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swampsnare Trap — Enchantment — Aura {2}{B}:
//
//	"This spell costs {1} less to cast if it targets a creature with
//	 flying.
//	 Enchant creature
//	 Enchanted creature gets -5/-3."
//
// The discount is Grow Extra Arms' self cost modifier over the Aura's
// announced enchant target (an Aura spell targets, CR 303.4a), judged
// as the spell is cast; flying is read off the creature as it is then,
// so a granted flying counts. It changes only the total cost, never the
// mana value (the ruling). The -5/-3 is an ordinary attached static.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "fc532d1d-6cf5-4b59-afb3-325d0bc0db77",
		Name:         "Swampsnare Trap",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(-5, -3)},
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a creature with flying.",
				And(Creature(), HasKeyword("flying"))),
		},
	})
}
