package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// First Responder — Creature — Ogre Citizen {3}{G}, 3/3 (Streets of New
// Capenna Commander):
//
//	"Vigilance
//	 At the beginning of your end step, you may return another creature
//	 you control to its owner's hand, then put a number of +1/+1 counters
//	 equal to that creature's power on this creature."
//
// Nothing targets, so the creature is a resolution-time pick (CR 608.2),
// zero being the "you may" decline, and the source itself is never
// offered (Ambrosia Whiteheart's shape). The power is read as the pick is
// made, while the creature is still on the battlefield, and carried to
// the tail that runs behind the return (its last-known information, CR
// 608.2h); "then" does not need the creature to reach a hand, so a
// commander that went to the command zone still counts. The counters go
// on only while this is still the same permanent on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8f5aed57-d0f3-4fa8-826c-0d94fa143f08",
		Name:            "First Responder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("First Responder — you may return another creature you control to its owner's hand, then put +1/+1 counters equal to its power on this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChooseAnotherCreatureYouControl(g, item,
						"First Responder — return another creature you control to its owner's hand?", firstResponderReturned)
				}),
		},
	})
}

// firstResponderReturned bounces the chosen creature and, behind the move,
// puts counters equal to the power it had on this creature.
var firstResponderReturned = effects.RegisterPermanentsPick("prompt-run/first-responder-return", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			tail := ctx.ItemTail(firstResponderCounters)
			tail.N = max(c.CurrentPower(), 0)
			return effects.BounceToHand{Target: id, ThenTail: tail}.Apply(ctx)
		}
		return nil
	},
})

// firstResponderCounters is "then put a number of +1/+1 counters equal to
// that creature's power on this creature". An on-disk key.
var firstResponderCounters = effects.CardTail("first-responder/counters", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: o.Tail.N}.Apply(ctx)
})
