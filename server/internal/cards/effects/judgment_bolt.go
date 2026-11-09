package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Judgment Bolt — Instant {3}{R} (Final Fantasy, #4462):
//
//	"Judgment Bolt deals 5 damage to target creature and X damage to
//	 that creature's controller, where X is the number of Equipment you
//	 control."
//
// Minion Missile's shape: the creature's controller is read off the
// target as it sits when the spell resolves (CR 608.2h), and the damage to
// them does not depend on the creature surviving. X is counted as the
// spell resolves, after the first 5 damage is dealt, and an X of 0 deals
// nothing. A target that has become illegal counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aad9f9f2-4c6b-48d0-80bd-77b73a9ea5f1",
		Name:         "Judgment Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, ref := range ctx.LegalTargets() {
				if ref.Kind != game.TargetCard {
					continue
				}
				victim, ok := ctx.Game.LookupCardForEffect(ref.ID)
				if !ok {
					return nil
				}
				if err := (DealDamage{Source: ctx.Source(), Target: ref.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
				x := countControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.HasSubtype("Equipment") })
				if x <= 0 {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: x}.Apply(ctx)
			}
			return nil
		},
	})
}
