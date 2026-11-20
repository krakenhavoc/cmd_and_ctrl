package effects

// Suture Spirit — Creature — Spirit {1}{W}, 1/1 (Eventide, #5590):
//
//	"Flying
//	 {W/B}{W/B}{W/B}: Regenerate target creature."
//
// The shield goes on the target as the ability resolves (CR 701.19a); a
// target gone by then gets nothing (lrwB1RegenerateTargetRow).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b2eba61-3937-4dc2-ad0a-e51fff8ecb80",
		Name:            "Suture Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			lrwB1RegenerateTargetRow("{W/B}{W/B}{W/B}: Regenerate target creature.", ManaCost("{W/B}{W/B}{W/B}"), TargetCreature("target creature")),
		},
	})
}
