package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narset's Rebuke — Instant {4}{R}:
//
//	"Narset's Rebuke deals 5 damage to target creature. Add {U}{R}{W}.
//	 If that creature would die this turn, exile it instead."
//
// In printed order: the damage, the mana (AddMana, into the caster's
// pool, emptying at the end of the step as all mana does), then the
// replacement, which is the spell's and marks the target whether or not
// the damage was dealt (ADR 0108 §1). A target that has become illegal
// counters the spell (CR 608.2b), so no mana is added either.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "247075c5-62f8-41a1-91b3-562ff0aabb30",
		Name:         "Narset's Rebuke",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			if err := (AddMana{Produced: "{U}{R}{W}"}).Apply(ctx); err != nil {
				return err
			}
			return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
		},
	})
}
