package effects

// Roast — Sorcery {1}{R} (Dragons of Tarkir,
// #5551):
//
//	"Roast deals 5 damage to target creature without flying."
//
// Charring Bite's body (Twinmaw Stormbrood).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5e60a85-2f31-47ec-a95b-2e69b2d2c00e",
		Name:         "Roast",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		Targets:      TargetCreature("target creature without flying", WithoutKeyword("flying")),
		OnResolve:    damageToFirstTarget(5),
	})
}
