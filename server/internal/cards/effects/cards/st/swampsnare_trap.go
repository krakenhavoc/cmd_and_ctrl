package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fc532d1d-6cf5-4b59-afb3-325d0bc0db77",
		Name:         "Swampsnare Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(-5, -3)},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a creature with flying.",
				effects.And(effects.Creature(), effects.HasKeyword("flying"))),
		},
	})
}
