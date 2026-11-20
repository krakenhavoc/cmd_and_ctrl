package effects

// Wilderness Hypnotist — Creature — Merfolk Wizard {2}{U}{U}, 1/3 (Eventide, #5590):
//
//	"{T}: Target red or green creature gets -2/-0 until end of turn."
//
// Humbling Elder's -2/-0 as a tap ability, over a two-colour clause read
// as it is announced and again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c6870f3-4b70-4050-bf22-b1f348f89360",
		Name:         "Wilderness Hypnotist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target red or green creature gets -2/-0 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -2, 0)),
			Cost:    TapCost(),
			Targets: TargetCreature("target red or green creature", Or(OfColor("R"), OfColor("G"))),
			Effect:  boostTargetUntilEOT("Wilderness Hypnotist — -2/-0 until end of turn", -2, 0),
		}},
	})
}
