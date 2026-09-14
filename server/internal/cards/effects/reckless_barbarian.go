package effects

// Reckless Barbarian — Creature — Dragon Barbarian {1}{R}, 2/2
// (EDHREC rank 3496):
//
//	"Sacrifice this creature: Add {R}{R}."
//
// A bear that becomes a ritual. The sacrifice is a mana ability's
// cost — Lotus Petal's shape — so it never uses the stack, and the
// mana lands before the dies-triggers the sacrifice queues resolve,
// which is what makes the Barbarian a storm card. No tap in the
// cost, so summoning sickness does not apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a621ae15-122a-4e11-bfcf-11dfc09784c0",
		Name:         "Reckless Barbarian",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Sacrifice: true},
			Produced: "{R}{R}",
			Label:    "Sacrifice this creature: Add {R}{R}",
		}},
	})
}
