package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// End-Blaze Epiphany — Instant {X}{R} (Lorwyn Eclipsed, #4458):
//
//	"End-Blaze Epiphany deals X damage to target creature. When that
//	 creature dies this turn, exile a number of cards from the top of your
//	 library equal to its power, then choose a card exiled this way. Until
//	 the end of your next turn, you may play that card."
//
// Blaze's X damage, at a creature, then a delayed triggered ability the
// spell creates as it resolves (Felonious Rage's whenThatCreatureDiesThisTurn,
// CR 603.7a), bound to that one object, so a creature that blinks and comes
// back is a new object whose death does nothing (CR 400.7). The power is the
// creature's as it last existed (CR 608.2h), the choice among the exiled
// cards is yours, and the chosen card alone gets the window (Heroes'
// Hangout's Date Night, with its own prompt), played with its usual costs and
// timing (the ruling). Cards you do not choose stay exiled. A target gone by
// resolution does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00cfbea0-e862-468b-90d5-7478eb9847c0",
		Name:         "End-Blaze Epiphany",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageXToTarget(0)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			return whenThatCreatureDiesThisTurn(ctx, id, "End-Blaze Epiphany — exile cards equal to its power, choose one to play", stdB66EndBlazeBody)
		},
	})
}
