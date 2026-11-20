package effects

// Engulfing Eruption — Sorcery {2}{R}{R} (Core Set 2020, #5518):
//
//	"Engulfing Eruption deals 5 damage to target creature."
//
// Direct Hit's body at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27b4fee0-ae46-4d15-96cd-aff8f3009de8",
		Name:         "Engulfing Eruption",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageToFirstTarget(5),
	})
}
