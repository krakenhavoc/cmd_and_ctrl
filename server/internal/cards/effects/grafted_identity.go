package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grafted Identity — Enchantment — Aura {2}{U}{U} (Innistrad: Midnight
// Hunt):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Enchant creature
//	 You control enchanted creature.
//	 Enchanted creature gets +1/+1."
//
// The cost is Altar's Reap's SacrificeCost, paid as the spell is cast
// (CR 601.2h); the control is Control Magic's layer-2 static and the pump
// is PumpAttached. Control reverts by itself when the Aura leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "e4fb97b5-0d4a-4c9a-bd51-42ff723b5854",
		Name:           "Grafted Identity",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        EnchantCreature(),
		Static: []game.StaticAbility{
			ControlAttachedBySource(),
			PumpAttached(1, 1),
		},
	})
}
