package effects

// Aetherjacket — Artifact Creature — Thopter {3}, 2/1 (Aetherdrift,
// #4464):
//
//	"Flying, vigilance
//	 {2}, {T}, Sacrifice this creature: Destroy another target artifact.
//	 Activate only as a sorcery."
//
// Westfold Rider's sacrifice-to-destroy at sorcery speed, with mana and a
// tap on the cost. "Another" is object identity (Another): the target is
// chosen before the cost is paid, so without it Aetherjacket,
// itself an artifact, could name itself. The target is re-checked at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "062c8e40-dea7-4df7-bcaa-f1d696fe61c3",
		Name:            "Aetherjacket",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice this creature: Destroy another target artifact. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets:      Another(TargetPermanent("another target artifact", Artifact())),
			SorcerySpeed: true,
			Effect:       destroyFirstLegalCardTarget,
		}},
	})
}
