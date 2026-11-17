package effects

// Adun Oakenshield — Legendary Creature — Human Knight {B}{R}{G}, 1/2
// (Legends, #5639):
//
//	"{B}{R}{G}, {T}: Return target creature card from your graveyard to
//	 your hand."
//
// A graveyard target clause (your graveyard, a creature card) and the
// shared returnTargetCardToHand body. With no creature card in your
// graveyard the ability has no legal target and can't be activated; a
// card that leaves the graveyard in response is no longer a legal
// target, and the ability does nothing (CR 608.2b). The tap means
// summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aefd49a4-ecec-42cb-b882-e5db7f81eff8",
		Name:         "Adun Oakenshield",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}{R}{G}, {T}: Return target creature card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{B}{R}{G}"), TapCost()),
			Targets: TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
			Effect:  returnTargetCardToHand,
		}},
	})
}
