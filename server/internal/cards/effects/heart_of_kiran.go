package effects

// Heart of Kiran — Legendary Artifact — Vehicle, 4/4, for {2}:
//
//	"Flying, vigilance
//	 Crew 3
//	 You may remove a loyalty counter from a planeswalker you control
//	 rather than pay Heart of Kiran's crew cost."
//
// SANDBOX SIMPLIFICATION, weaker than printed: the alternative crew
// cost is not here. "Rather than pay" is an alternative COST on an
// activated ability, and game.AbilityCost has no alternative-cost
// slot — the S22 AlternativeCost machinery is a CAST-time clause
// (CR 118.9) that only spells reach. Building a second, cheaper
// Activated entry to fake it would be worse than leaving it out: it
// would be activatable with no planeswalker in play at all, which is
// the stronger-than-printed direction #259 exists to forbid.
//
// So Heart of Kiran ships as a strictly worse Heart of Kiran — crew
// 3 or nothing — and the day AbilityCost grows an alternative-cost
// component, this card is one line of the fix.
func init() {
	Register(Spec{
		OracleID:        "e2ee410f-2467-4f1f-84a0-8a79faedc0b3",
		Name:            "Heart of Kiran",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"You can't remove a loyalty counter from a planeswalker to crew it; only Crew 3 is offered."},
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []ActivatedAbility{{
			Label:  "Crew 3",
			Cost:   CrewCost(3),
			Effect: CrewEffect("Heart of Kiran"),
		}},
	})
}
