package effects

// Fell the Pheasant — Instant {1}{G} (Throne of Eldraine, #5516):
//
//	"Fell the Pheasant deals 5 damage to target creature with flying. Create a Food token."
//
// The damage, then the Food as a second step (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c78c8f7d-45fd-46c0-9ed4-4d0e85aada2e",
		Name:         "Fell the Pheasant",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		Targets:      TargetCreature("target creature with flying", HasKeyword("flying")),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return sourceDealsDamageToEachLegalTarget(5)(ctx.Game, ctx.Item) }),
			CreateToken{Template: FoodToken(), N: 1},
		),
	})
}
