package effects

// Blood Pet — Creature — Thrull {B}, 1/1 (EDHREC rank 2638):
//
//	"Sacrifice this creature: Add {B}."
//
// The one-drop that is a Dark Ritual's worth of tempo across two
// turns, and a body for an aristocrats deck to feed to itself. The
// ability is a mana ability with a sacrifice-this cost (Lotus
// Petal's shape): it never uses the stack, the {B} lands in the pool
// at once, and the Thrull's death fires every dies-trigger and
// "whenever you sacrifice" payoff after the mana is already there
// (CR 605.3a). No tap in the cost, so summoning sickness does not
// apply — it can be cracked the turn it enters, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e05c6c80-a91a-45e0-b991-0014fd5a6472",
		Name:         "Blood Pet",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Sacrifice: true},
			Produced: "{B}",
			Label:    "Sacrifice this creature: Add {B}",
		}},
	})
}
