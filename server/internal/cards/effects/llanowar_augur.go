package effects

// Llanowar Augur — Creature — Elf Shaman {G}, 0/3 (Future Sight, #5594):
//
//	"Sacrifice this creature: Target creature gets +3/+3 and gains
//	 trample until end of turn. Activate only during your upkeep."
//
// An upkeep-only sacrifice (DuringYourUpkeep is the step AND your turn)
// with the shared pump-and-grant body on its one target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3052531-f7cf-44a7-9492-eb155d75268c",
		Name:         "Llanowar Augur",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Sacrifice this creature: Target creature gets +3/+3 and gains trample until end of turn. Activate only during your upkeep.",
			Purpose:   ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
			Cost:      SacrificeThis(),
			Condition: DuringYourUpkeep(),
			Targets:   TargetCreature("target creature"),
			Effect:    pumpAndGrantFirstTarget(3, 3, "Llanowar Augur — +3/+3 and trample until end of turn", "trample"),
		}},
	})
}
