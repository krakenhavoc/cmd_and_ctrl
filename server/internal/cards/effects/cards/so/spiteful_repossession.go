package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiteful Repossession — Sorcery {4}{R} (Streets of New Capenna
// Commander):
//
//	"Spiteful Repossession deals damage to each opponent who controls
//	 more lands than you equal to the difference. Then create a number of
//	 Treasure tokens equal to the damage dealt this way."
//
// Each opponent's difference is fixed as the spell resolves, in APNAP
// order (CR 101.4); an opponent with no more lands than you takes
// nothing. The damage is dealt to one opponent at a time through the
// keyed damage continuation, which carries the opponents still to go,
// their amounts and the damage dealt so far on the tail, so a prompt
// over one opponent's damage is a restore point. The Treasures are as
// many as the damage that was actually dealt: damage that was prevented
// makes none. They are created once the last opponent is done.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63fd77e2-348a-4f6c-88b4-465fbe339d29",
		Name:         "Spiteful Repossession",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			mine := len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Land()))
			var opps []uuid.UUID
			var amounts []int
			for _, opp := range effects.ApnapOrder(ctx.Game, ctx.Opponents()) {
				if d := len(effects.PermanentsControlledByMatching(ctx.Game, opp, effects.Land())) - mine; d > 0 {
					opps = append(opps, opp)
					amounts = append(amounts, d)
				}
			}
			t := ctx.Tail(spitefulRepossessionStep)
			t.Carry, t.Nums, t.N = opps, amounts, 0
			return spitefulRepossessionNext(ctx.Game, t)
		},
	})
}

// spitefulRepossessionStep is told the damage the last opponent took (its
// Amount), adds it to the running total on the tail and goes on. An
// on-disk key.
var spitefulRepossessionStep = game.RegisterTail("then/spiteful-repossession-step", func(g *game.Game, o game.TailOutcome) error {
	t := o.Tail
	t.N += o.Amount
	return spitefulRepossessionNext(g, t)
})

// spitefulRepossessionNext deals the damage of the next opponent on the
// tail, or, with none left, creates a Treasure for each point dealt.
func spitefulRepossessionNext(g *game.Game, t game.KeyedTail) error {
	if len(t.Carry) == 0 || len(t.Nums) == 0 {
		if t.N <= 0 {
			return nil
		}
		return effects.CreateToken{Controller: t.Player, Template: effects.RuleDefinedToken(effects.TreasureToken()), N: t.N}.Apply(effects.TailContext(g, t))
	}
	opp, amount := t.Carry[0], t.Nums[0]
	t.Carry, t.Nums = append([]uuid.UUID(nil), t.Carry[1:]...), append([]int(nil), t.Nums[1:]...)
	return g.DealDamageToPlayerKeyedForEffect(t.Source, opp, amount, t)
}
