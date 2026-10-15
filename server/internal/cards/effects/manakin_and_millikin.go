package effects

// Manakin and Millikin — Artifact Creature — Construct {3}, 1/2:
//
//	"{T}, Mill a card: Add {C}{C}."
//
// Millikin's ability for twice the mana. The mill is a real cost
// component (ManaAbilityCost.Mill, #3845, ADR 0020's 2026-10-15
// amendment), validated with the tap and paid last (CR 601.2h), so an
// empty library can't pay it and the ability is not offered
// (CR 701.17b). A mana ability never uses the stack (CR 605.3b), and
// the {T} is on a creature, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ceda813a-7c15-47a9-a8a3-4724f1dfbf35",
		Name:         "Manakin and Millikin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mill: 1},
			Produced: "{C}{C}",
			Label:    "{T}, Mill a card: Add {C}{C}",
		}},
	})
}
