package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloatfly Swarm — Creature — Insect Mutant {3}{B}, 0/0:
//
//	"Flying
//	 This creature enters with five +1/+1 counters on it.
//	 If damage would be dealt to this creature while it has a +1/+1
//	 counter on it, prevent that damage, remove that many +1/+1 counters
//	 from it, then give each player a rad counter for each +1/+1 counter
//	 removed this way."
//
// ADR 0108 §8 (#1906) is the prevention static and its additional
// effect (CR 615.5), Ugin's Conjurant's shape: "that many" is the damage,
// prevented or not (CR 615.12), or every counter it has when that is
// fewer. #2042 made the rad counters do something (CR 728.1), which is
// what held the card back. The rad counters follow the counters that
// actually came off, read before and after the removal, and the Swarm's
// controller gives them to every player still in the game, themselves
// included.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc50e254-71bc-46b3-93e6-c6eb820a3473",
		Name:            "Bloatfly Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 5, "Bloatfly Swarm: enters with five +1/+1 counters"),
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				While: effects.WhileItHasAPlusOneCounter,
				Then:  bloatflyRadBody,
				Label: "Bloatfly Swarm — prevent damage to it, remove that many +1/+1 counters and give each player that many rad counters",
			}),
		},
	})
}

// bloatflyRadBody is "remove that many +1/+1 counters from it, then give
// each player a rad counter for each +1/+1 counter removed this way". An
// on-disk key: never rename it.
var bloatflyRadBody = game.DelayedBody("prevention/bloatfly-remove-counters-give-rad", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	info, ok := effects.FollowUpThis(g, item)
	if !ok {
		return nil
	}
	before := info.Counters[game.CounterPlusOne]
	n := min(effects.ThatDamage(item), before)
	if n <= 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	ctx.RecordStep(bloatflyRecord, nil, before, n)
	return bloatflyRemoveThenRad.Run(ctx)
})

// bloatflyRecord is the +1/+1 counters it had and how many are to come off.
const bloatflyRecord = "bloatfly-swarm/counters"

// bloatflyRemoveThenRad takes the counters off and then gives the rad counters:
// they wait for a prompt the removal asks (a CR 616 order; CR 608.2c, #4155)
// and count how many really came off.
var bloatflyRemoveThenRad = effects.RegisterThenSteps("then-steps/bloatfly-swarm-remove",
	effects.Instruction(func(ctx *effects.Context) error {
		return ctx.Game.AddCounterForEffect(ctx.Item.SourceCardID, game.CounterPlusOne, -ctx.StepRecord(bloatflyRecord).Ns[1])
	}),
	effects.Instruction(func(ctx *effects.Context) error {
		r := ctx.StepRecord(bloatflyRecord)
		removed := r.Ns[1]
		if after, ok := effects.FollowUpThis(ctx.Game, ctx.Item); ok {
			removed = r.Ns[0] - after.Counters[game.CounterPlusOne]
		}
		return effects.EachPlayerGetsRadCounters(ctx.Game, ctx.Controller(), removed)
	}),
)
