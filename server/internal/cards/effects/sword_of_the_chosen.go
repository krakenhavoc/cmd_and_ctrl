package effects

// Sword of the Chosen — Legendary Artifact {2} (Stronghold, #5627):
//
//	"{T}: Target legendary creature gets +2/+2 until end of turn."
//
// Willow Satyr's "target legendary creature" clause and Angelic Page's
// shared boostTargetUntilEOT body at +2/+2, ending in the cleanup step
// (CR 514.2). It is an artifact that isn't a creature, so the tap can be
// paid the turn it arrives (CR 302.6 applies to creatures only). It is
// not Equipment and attaches to nothing. A target that stops being
// legendary in response is no longer legal and gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c756eda-4fc1-4766-99be-32d8c9f35262",
		Name:         "Sword of the Chosen",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target legendary creature gets +2/+2 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 2, 2)),
			Cost:    TapCost(),
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect:  boostTargetUntilEOT("Sword of the Chosen — +2/+2", 2, 2),
		}},
	})
}
