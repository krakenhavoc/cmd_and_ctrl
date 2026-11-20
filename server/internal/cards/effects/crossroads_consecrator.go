package effects

// Crossroads Consecrator — Creature — Human Cleric {G}, 1/2 (Eldritch
// Moon, #5541):
//
//	"{G}, {T}: Target attacking Human gets +1/+1 until end of turn."
//
// Infantry Veteran's pump at {G}, narrowed to an attacking Human.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4f41c7b-b4ae-46cd-9fdc-8b8e6c29a62d",
		Name:         "Crossroads Consecrator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: Target attacking Human gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Targets: TargetCreature("target attacking Human", AttackingCreature(), HasSubtype("Human")),
			Effect:  pumpAndGrantFirstTarget(1, 1, "Crossroads Consecrator — +1/+1 until end of turn"),
		}},
	})
}
