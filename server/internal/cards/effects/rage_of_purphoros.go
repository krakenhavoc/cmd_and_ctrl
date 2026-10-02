package effects

// Rage of Purphoros — Sorcery {4}{R}:
//
//	"Rage of Purphoros deals 4 damage to target creature. It can't be
//	 regenerated this turn. Scry 1."
//
// The no-regeneration rider is the spell's, not the damage's (ADR 0108
// §2). The scry happens after both, as printed, and only if the target is
// still legal (an illegal target means the spell does not resolve at all).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d6667a93-c706-451f-a83d-ffe5b9b9e53e",
		Name:         "Rage of Purphoros",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: damageFirstTargetThenNoRegen(4, func(ctx *Context) error {
			return Scry{N: 1}.Apply(ctx)
		}),
	})
}
