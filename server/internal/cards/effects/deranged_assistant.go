package effects

// Deranged Assistant — Creature — Human Wizard {1}{U}, 1/1:
//
//	"{T}, Mill a card: Add {C}. (Activate only as an instant. To mill
//	 a card, put the top card of your library into your graveyard.)"
//
// Millikin's ability on a blue body: the printed cost is a tap and a
// mill, both real cost components (ManaAbilityCost.Mill, #3845, ADR
// 0020's 2026-10-15 amendment). The mill is validated with the tap, so
// an empty library can't pay it and the ability is not offered
// (CR 701.17b), and it is paid last (CR 601.2h) through the ordinary
// mill, so Rest in Peace exiles the card instead and a "whenever you
// mill" payoff sees it. A mana ability never uses the stack
// (CR 605.3b), and the {T} is on a creature, so summoning sickness
// applies (CR 302.6), which the engine enforces from the type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e1b0034-bd2c-4412-8d81-db3306c8814f",
		Name:         "Deranged Assistant",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mill: 1},
			Produced: "{C}",
			Label:    "{T}, Mill a card: Add {C}",
		}},
	})
}
