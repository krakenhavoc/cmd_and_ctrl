package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const honestWorkGrant = "honest-work/tap-for-colorless"

// Honest Work — Enchantment — Aura {U} (Avatar: The Last Airbender,
// #4459):
//
//	"Enchant creature an opponent controls
//	 When this Aura enters, tap enchanted creature and remove all
//	 counters from it.
//	 Enchanted creature loses all abilities and is a Citizen with base
//	 power and toughness 1/1 and '{T}: Add {C}' named Humble Merchant.
//	 (It loses all other creature types and names.)"
//
// The entry trigger is Enchanted River's Grasp's (one shared body). The
// static is one effect across four layers, each in the layer the rules
// put it:
//
//   - layer 3: named Humble Merchant (CR 612.8, a text-changing
//     effect, not a copy effect);
//   - layer 4: its creature types become Citizen alone; card types,
//     supertypes and non-creature subtypes stay;
//   - layer 6: loses all abilities and has "{T}: Add {C}", one effect,
//     so the granted mana ability survives the removal (Imprisoned in
//     the Moon's shape);
//   - layer 7b: base power and toughness 1/1, so counters and pumps
//     still apply on top.
//
// Abilities gained after the Aura's timestamp still apply, as the ruling
// says, because layer 6 applies effects in timestamp order.
//
// No simplification.
func init() {
	removal := LoseAllAbilities()
	removal.GrantAbilities = []string{honestWorkGrant}
	Register(Spec{
		OracleID:     "12956c07-e679-471b-8dcb-2a2d15b0be55",
		Name:         "Honest Work",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(OpponentControls()),
		Grants: []AbilityGrant{
			TapForManaGrant(honestWorkGrant, "{C}", "Add {C}", "{T}: Add {C}."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Honest Work — tap enchanted creature and remove all counters from it",
				tapEnchantedAndRemoveAllCounters),
		},
		Static: []game.StaticAbility{
			avB4SetAttachedName("Humble Merchant"),
			SetAttachedCreatureTypes("Citizen"),
			removal,
			SetAttachedBasePT(1, 1),
		},
	})
}
