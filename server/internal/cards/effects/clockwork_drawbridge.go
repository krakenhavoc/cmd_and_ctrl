package effects

// Clockwork Drawbridge — Artifact Creature — Wall {W}, 0/3 (Dominaria United, S82):
//
//	"Defender
//	 {2}{W}, {T}: Tap target creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e49eb21c-2a0d-4fd0-9f69-1bff6e94f9e7",
		Name:            "Clockwork Drawbridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
