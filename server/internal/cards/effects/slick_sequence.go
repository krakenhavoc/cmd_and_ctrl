package effects

import "github.com/google/uuid"

// Slick Sequence — Instant {U}{R} (Outlaws of Thunder Junction, #4468):
//
//	"Slick Sequence deals 2 damage to any target. If you've cast another
//	 spell this turn, draw a card."
//
// Shock's damage, then a draw gated on the controller's cast tally as
// the spell resolves. The tally counts Slick Sequence itself when it was
// cast, so "another spell" is one more than that; a copy was never cast
// (CR 707.10), so for a copy any spell in the tally is another one. The
// draw waits behind the damage's own prompts (a CR 616 order, CR
// 608.2c). With its only target gone by resolution the spell does
// nothing at all, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ac8c118-3942-4c92-856c-20d764283650",
		Name:         "Slick Sequence",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				ctx.RecordStep(slickSequenceRecord, []uuid.UUID{legal[0].ID})
				return DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: 2}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if len(ctx.StepRecord(slickSequenceRecord).IDs) == 0 {
					return nil
				}
				others := ctx.Game.CastTallyFor(ctx.Controller()).Total
				if !ctx.Item.IsCopy {
					others--
				}
				if others <= 0 {
					return nil
				}
				return DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}

// slickSequenceRecord files the target the damage was dealt to.
const slickSequenceRecord = "slick-sequence/target"
