package effects

// Ridged Kusite — Creature — Horror Spellshaper {B}, 1/1 (Planar Chaos, #5595):
//
//	"{1}{B}, {T}, Discard a card: Target creature gets +1/+0 and gains first
//	 strike until end of turn."
//
// pumpAndGrantFirstTarget on an activated row whose cost discards: the
// still-legal target (CR 608.2b) gets the boost and the keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0c3c26e-e9e8-424d-922e-a692bb479ebd",
		Name:         "Ridged Kusite",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}, Discard a card: Target creature gets +1/+0 and gains first strike until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")),
			Cost:    Plus(ManaCost("{1}{B}"), TapCost(), DiscardACard()),
			Targets: TargetCreature("target creature"),
			Effect:  pumpAndGrantFirstTarget(1, 0, "Ridged Kusite — +1/+0 and first strike until end of turn", "first strike"),
		}},
	})
}
