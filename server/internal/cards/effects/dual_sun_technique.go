package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dual-Sun Technique — Instant {1}{W} (Edge of Eternities, #4461):
//
//	"Target creature you control gains double strike until end of turn. If
//	 it has a +1/+1 counter on it, draw a card."
//
// The grant, then the check, in printed order (CR 608.2c): the counter is
// read as the spell resolves, so a counter put on the creature in response
// still draws, and one removed in response does not. A creature that left
// in response is an illegal target and the spell does nothing, the draw
// included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89ec0f03-47d0-4657-8987-95e680404f24",
		Name:         "Dual-Sun Technique",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "double strike"))},
		Targets: TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"},
				Label: "Dual-Sun Technique — double strike until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || c.Counters[game.CounterPlusOne] <= 0 {
				return nil
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
