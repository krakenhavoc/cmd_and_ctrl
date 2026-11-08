package effects

// Winter's Intervention — Instant {1}{B} (Duskmourn: House of Horror,
// #4466):
//
//	"Winter's Intervention deals 2 damage to target creature. You gain 2
//	 life."
//
// Defibrillating Current's two steps on a creature target: the damage,
// then the life, each its own step so the gain waits for any prompt the
// damage asks (CR 608.2c). A target that is gone as the spell resolves
// fizzles the whole spell, life included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91967284-6711-4ef3-bb9b-5ba873da8bb2",
		Name:         "Winter's Intervention",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, ref := range ctx.LegalTargets() {
						if err := (DealDamage{Source: ctx.Item.SourceCardID, Target: ref.ID, Amount: 2}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Item.Controller, Amount: 2}.Apply(ctx)
			}),
		),
	})
}
