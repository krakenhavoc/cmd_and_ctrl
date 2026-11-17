package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unblinking Observer — Creature — Homunculus {1}{U}, 2/1 (Innistrad:
// Midnight Hunt, #5680):
//
//	"{T}: Add {U}. Spend this mana only to pay a disturb cost or cast an
//	 instant or sorcery spell."
//
// One mana ability (CR 605.1a) whose mana carries one any-of tag (CR
// 106.6): a cast that claims the disturb offer (ManaRestrictAltCost, read
// off the key Disturb writes, so the two cannot drift), or a cast of an
// instant or sorcery spell (Hydro-Channeler's tags). A disturb cast's
// total cost is the disturb cost with its increases and reductions (CR
// 601.2f), and the mana pays any of it. It does not pay for the same
// card cast from hand for its mana cost, for a graveyard cast under any
// other alternative cost (an escaped creature), or for an activation.
// Summoning sickness gates the tap (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     unblinkingObserverOracleID,
		Name:         "Unblinking Observer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictAltCost(Disturb("").Key)},
				[]string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
			)},
			Label: "Add {U}. Spend this mana only to pay a disturb cost or cast an instant or sorcery spell",
		}},
	})
}

const unblinkingObserverOracleID = "871c4b27-990a-4947-b962-8a6d0bfd85ad"
