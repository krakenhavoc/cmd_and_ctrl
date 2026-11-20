package effects

// Ghost Warden — Creature — Spirit {1}{W}, 1/1 (Guildpact, #5599):
//
//	"{T}: Target creature gets +1/+1 until end of turn."
//
// Wyluli Wolf's ability, written with Icatian Priest's shared body
// (pumpAndGrantFirstTarget): the still-legal target gets the boost,
// nothing if it left in response (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66bd5999-742a-43e0-827c-ecb8ec7db242",
		Name:         "Ghost Warden",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  pumpAndGrantFirstTarget(1, 1, "Ghost Warden — +1/+1 until end of turn"),
		}},
	})
}
