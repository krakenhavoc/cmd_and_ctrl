package effects

// Scorching Shot — Sorcery {R}{R} (Outlaws of Thunder Junction, #4468):
//
//	"Scorching Shot deals 5 damage to target creature."
//
// Direct Hit's body (the timing is the card type's). A target that is
// gone or no longer a creature as it resolves takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e13113ba-1923-493c-a48d-ca320b7adbde",
		Name:         "Scorching Shot",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageToFirstTarget(5),
	})
}
