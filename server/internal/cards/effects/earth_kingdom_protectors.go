package effects

// Earth Kingdom Protectors — Creature — Human Soldier Ally {W}, 1/1
// (Avatar: The Last Airbender, #4459):
//
//	"Vigilance
//	 Sacrifice this creature: Another target Ally you control gains
//	 indestructible until end of turn."
//
// "Another" is object identity (Another, ExcludeSource): the Protectors
// can never pick themselves, and are gone by the time the ability
// resolves anyway, since the sacrifice is the cost. The Ally is any
// permanent you control with the type, read post-layer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3bfb3913-808d-444b-97c3-a90af3bc585b",
		Name:            "Earth Kingdom Protectors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Another target Ally you control gains indestructible until end of turn.",
			Cost:    SacrificeThis(),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")),
			Targets: Another(TargetPermanent("another target Ally you control", HasSubtype("Ally"), YouControl())),
			Effect:  firstTargetGainsUntilEndOfTurn(0, "Earth Kingdom Protectors — indestructible until end of turn", "indestructible"),
		}},
	})
}
