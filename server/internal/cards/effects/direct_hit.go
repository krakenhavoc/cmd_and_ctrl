package effects

// Direct Hit — Sorcery {2}{R} (Star Trek Commander, #4974):
//
//	"Direct Hit deals 5 damage to target creature."
//
// Concentrated Fire's body at sorcery speed (the timing is the card
// type's). A target that is gone or no longer a creature as it resolves
// takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6b0adcb-4676-4d9f-83ae-a2ff3cdfd84b",
		Name:         "Direct Hit",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageToFirstTarget(5),
	})
}
