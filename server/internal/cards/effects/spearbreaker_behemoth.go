package effects

// Spearbreaker Behemoth — Creature — Beast {5}{G}{G}, 5/5 (Reprint, #5908):
//
//	"Indestructible
//	 {1}: Target creature with power 5 or greater gains indestructible
//	 until end of turn."
//
// The power condition is part of the target (CR 115.1), judged at
// announcement and again at resolution, so a creature that shrank in
// response is an illegal target and the ability does nothing (CR 608.2b).
// The grant is Burdened Stoneback's, a layer-6 ability addition that ends
// at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "025f8db8-11d1-4c59-8e1f-004c9c843fd8",
		Name:            "Spearbreaker Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Activated: []ActivatedAbility{{
			Label:   "{1}: Target creature with power 5 or greater gains indestructible until end of turn.",
			Cost:    ManaCost("{1}"),
			Targets: TargetCreature("target creature with power 5 or greater", PowerGE(5)),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")),
			Effect:  ebTargetCreatureGainsUntilEOT("indestructible", "Spearbreaker Behemoth — indestructible"),
		}},
	})
}
