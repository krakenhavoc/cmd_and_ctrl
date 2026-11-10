package effects

// Incinerating Blast — Sorcery {4}{R} (Foundations, #4465):
//
//	"Incinerating Blast deals 6 damage to target creature.
//	 You may discard a card. If you do, draw a card."
//
// Joust Through's damage step at 6, then Flaring Cinder's rummage
// (b39MayDiscardThenDraw): the discard is optional, and a card is drawn
// only for a card discarded. Each is its own step, so the rummage waits for
// any prompt the damage asks (CR 608.2c). A target that is gone as the
// spell resolves fizzles the whole spell, rummage included (CR 608.2b).
//
// No simplification.
func init() {
	const rummage = "Incinerating Blast — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "49f0da4a-e046-4220-b90e-e105a5f2ef00",
		Name:         "Incinerating Blast",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 6)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, ref := range ctx.LegalTargets() {
						if err := (DealDamage{Source: ctx.Item.SourceCardID, Target: ref.ID, Amount: 6}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			Instruction(b39MayDiscardThenDraw(1, false, rummage, drawOneIfDiscarded)),
		),
	})
}
