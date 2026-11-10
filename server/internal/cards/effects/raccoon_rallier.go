package effects

// Raccoon Rallier — Creature — Raccoon Bard {1}{R}, 2/2 (Bloomburrow,
// #4467):
//
//	"{T}: Target creature you control gains haste until end of turn.
//	 Activate only as a sorcery."
//
// A tap ability, so summoning sickness applies to the Rallier itself
// (the engine enforces it for a creature's {T} cost). The target is
// re-checked as the ability resolves, and the haste lasts for the turn.
//
// No simplification.
func init() {
	const label = "Raccoon Rallier — haste until end of turn"
	Register(Spec{
		OracleID:     "50c038a4-6546-4a15-a117-176849edaa3d",
		Name:         "Raccoon Rallier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: Target creature you control gains haste until end of turn. Activate only as a sorcery.",
			Cost:         TapCost(),
			Targets:      TargetCreature("target creature you control", YouControl()),
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			SorcerySpeed: true,
			Effect:       firstTargetGainsUntilEndOfTurn(0, label, "haste"),
		}},
	})
}
