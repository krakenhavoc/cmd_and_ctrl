package effects

// Angelic Page — Creature — Angel Spirit {1}{W}, 1/1 (Urza's Saga, #5624):
//
//	"Flying
//	 {T}: Target attacking or blocking creature gets +1/+1 until end of
//	 turn."
//
// Alaborn Veteran's pump body (boostTargetUntilEOT) at instant speed, on
// a target that must be attacking or blocking when it is chosen and again
// as the ability resolves (CR 608.2b): a creature removed from combat in
// response gets nothing. Any such creature, yours or an opponent's. The
// tap cost keeps it from being activated the turn it arrives (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7f76abc-8656-4e97-9b7f-7f89ca3e6d93",
		Name:            "Angelic Page",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target attacking or blocking creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Effect:  boostTargetUntilEOT("Angelic Page — +1/+1", 1, 1),
		}},
	})
}
