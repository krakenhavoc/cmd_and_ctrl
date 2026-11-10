package effects

// Joust Through — Instant {W} (Foundations, #4465):
//
//	"Joust Through deals 3 damage to target attacking or blocking
//	 creature. You gain 1 life."
//
// Winter's Intervention's two steps on Elspeth's Smite's target: the
// damage, then the life, each its own step so the gain waits for any
// prompt the damage asks (CR 608.2c). A target that is gone, or no longer
// attacking or blocking, as the spell resolves fizzles the whole spell,
// life included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33c21ef1-33eb-447a-a318-a352f163a8a0",
		Name:         "Joust Through",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, ref := range ctx.LegalTargets() {
						if err := (DealDamage{Source: ctx.Item.SourceCardID, Target: ref.ID, Amount: 3}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Item.Controller, Amount: 1}.Apply(ctx)
			}),
		),
	})
}
