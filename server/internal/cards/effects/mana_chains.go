package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Chains — Enchantment — Aura {U}:
//
//	"Enchant creature
//	 Enchanted creature has 'Cumulative upkeep {1}.'"
//
// #3869. A layer-6 grant of the CumulativeUpkeepGrant bundle to the
// enchanted creature (ADR 0093). The trigger is the creature's: its
// controller is asked at their upkeep, the age counters go on the
// creature, and a decline sacrifices the creature. A creature with its
// own cumulative upkeep, or under two Mana Chains, triggers once per
// instance, and each instance counts every age counter (CR 702.24b).
// The counters stay if the Aura leaves; the ability does not.
//
// No simplification.
const manaChainsGrant = "mana-chains/cumulative-upkeep-1"

func init() {
	Register(Spec{
		OracleID:     "340d27c6-2e8e-472a-85f5-9b36d9cc3e42",
		Name:         "Mana Chains",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants:       []AbilityGrant{CumulativeUpkeepGrant(manaChainsGrant, "{1}")},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesToAttached(manaChainsGrant), "Enchanted creature has \"Cumulative upkeep {1}.\""),
		},
	})
}
