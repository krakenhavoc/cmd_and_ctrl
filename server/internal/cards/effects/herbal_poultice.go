package effects

// Herbal Poultice — Artifact {0} (Lorwyn, #5593):
//
//	"{3}, Sacrifice this artifact: Regenerate target creature."
//
// The artifact goes as part of the cost (CR 701.21a), so the shield lands
// even if the Poultice is gone by resolution. It replaces the target's next
// destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a8e1f8e-3740-4ff3-b7fa-390ea278fb95",
		Name:         "Herbal Poultice",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			lrwB1RegenerateTargetRow("{3}, Sacrifice this artifact: Regenerate target creature.",
				Plus(ManaCost("{3}"), SacrificeThis()), TargetCreature("target creature")),
		},
	})
}
