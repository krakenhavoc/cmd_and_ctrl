package effects

// Grandmother Sengir — Legendary Creature — Human Wizard {4}{B}, 3/3
// (Homelands, #5635):
//
//	"{1}{B}, {T}: Target creature gets -1/-1 until end of turn."
//
// The {T} keeps her off the turn she arrives (CR 602.5a). The change is
// layer 7c (CR 613.4c) and ends at cleanup (CR 514.2); a target that
// left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f92804f-a231-4406-9a16-f4e102517b06",
		Name:         "Grandmother Sengir",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}: Target creature gets -1/-1 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -1, -1)),
			Cost:    Plus(ManaCost("{1}{B}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Grandmother Sengir — -1/-1 until end of turn", -1, -1),
		}},
	})
}
