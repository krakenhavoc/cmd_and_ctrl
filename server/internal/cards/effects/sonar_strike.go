package effects

// Sonar Strike — Instant {1}{W} (Bloomburrow, #4467):
//
//	"Sonar Strike deals 4 damage to target attacking, blocking, or tapped
//	 creature. You gain 3 life if you control a Bat."
//
// The target is judged at announce and again as the spell resolves
// (CR 608.2b): a creature that is no longer attacking, blocking or
// tapped counters the spell, and then no life is gained either. "If you
// control a Bat" is read as the spell resolves, through the spell's own
// text (ctx.TextSubtype). Two steps, so the life waits for any question
// the damage asks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8c582bc-47de-47ec-8bf9-b2c848e27ab8",
		Name:         "Sonar Strike",
		Completeness: CompletenessFull,
		Targets: TargetCreature("target attacking, blocking, or tapped creature",
			Or(AttackingOrBlocking(), tappedPermanent())),
		Purpose: ForTargets(DamageToTarget(0, 4)),
		OnResolve: ResolveSteps(
			damageTheClauseTarget(0, 4),
			Instruction(func(ctx *Context) error {
				if !ControlsA(ctx.TextSubtype("Bat"))(ctx.Game, ctx.Controller()) {
					return nil
				}
				return GainLife{Amount: 3}.Apply(ctx)
			}),
		),
	})
}
