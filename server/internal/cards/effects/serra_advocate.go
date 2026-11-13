package effects

// Serra Advocate — Creature — Angel {3}{W}, 2/2 (Urza's Destiny, #5621):
//
//	"Flying
//	 {T}: Target attacking or blocking creature gets +2/+2 until end of
//	 turn."
//
// Angelic Page's row at +2/+2: the target must be attacking or blocking
// as the ability is activated and again as it resolves (CR 608.2b), so
// it is a combat trick; the boost ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c67292a-7fcc-4646-8a06-99a7b4b3e005",
		Name:            "Serra Advocate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target attacking or blocking creature gets +2/+2 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 2, 2)),
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Effect:  boostTargetUntilEOT("Serra Advocate — +2/+2", 2, 2),
		}},
	})
}
