package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Katara's Reversal — Instant {2}{U}{U} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Counter up to four target spells and/or abilities.
//	 Untap up to four target artifacts and/or creatures."
//
// Two target clauses, each up to four: the first takes spells and
// activated or triggered abilities on the stack (Disallow's clause), the
// second artifacts and creatures on the battlefield. Each pick still
// legal as it resolves is countered or untapped (CR 608.2b); a pick that
// became illegal is left alone, and the spell does nothing only if every
// pick is illegal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83ccbcb1-7cf0-41a0-8abd-18f4bc647b5f",
		Name:         "Katara's Reversal",
		Completeness: CompletenessFull,
		Targets: TargetSpellOrAbility("up to four target spells and/or abilities").WithCount(0, 4).Then(
			TargetPermanent("up to four target artifacts and/or creatures", Or(Artifact(), Creature())).WithCount(0, 4)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.ClauseTargets(0) {
				if !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range ctx.ClauseTargets(1) {
				if !ctx.IsTargetLegal(t) || t.Kind != game.TargetCard {
					continue
				}
				if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
