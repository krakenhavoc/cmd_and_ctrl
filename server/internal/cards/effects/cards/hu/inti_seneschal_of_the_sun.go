package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inti, Seneschal of the Sun — Legendary Creature — Human Knight {1}{R},
// 2/2 (The Lost Caverns of Ixalan, #4470):
//
//	"Whenever you attack, you may discard a card. When you do, put a
//	 +1/+1 counter on target attacking creature. It gains trample until
//	 end of turn.
//	 Whenever you discard one or more cards, exile the top card of your
//	 library. You may play that card until your next end step."
//
// "Whenever you attack" is ONE trigger per attack declaration
// (OncePerBatch, Banon's shape). The discard is the trigger's "you may",
// asked as it resolves through MayPay with a discard payment (#4566), so
// a player with an empty hand is not asked. Only a card really discarded
// raises the CR 603.12 reflexive trigger, which picks its attacking
// creature as it goes on the stack and, if that creature is still legal,
// puts the counter on it and gives it trample until end of turn as one
// effect (untilEndOfTurn).
//
// The second trigger is Cryptcaller Chariot's discard watch, held to the
// end of the discard (ThatManyEvents) so it goes on the stack once per
// batch however many cards were in it; the count itself is not used. It
// exiles one card and lets you play it until your next end step begins
// (ExileTopNUntilYourNextEndStep). The first trigger's own discard sets
// it off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "176f5ba8-6e7b-4234-a298-b6164f51656f",
		Name:         "Inti, Seneschal of the Sun",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Inti, Seneschal of the Sun — you may discard a card", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Pay:      effects.DiscardPayment(1),
					Question: "Inti — discard a card to put a +1/+1 counter on target attacking creature and give it trample?",
					Then:     intiDiscarded,
				}.Apply(effects.NewContext(g, item))
			})),
			effects.WithAtBatchEnd(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Inti, Seneschal of the Sun — exile the top card of your library; you may play it until your next end step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopNUntilYourNextEndStep(effects.NewContext(g, item), 1)
				}), effects.ThatManyEvents()),
		},
	})
}

// intiDiscarded is the "when you do": the reflexive trigger, raised only
// once the card was really discarded. Keyed: never renamed, never reused.
var intiDiscarded = effects.PayUnlessBranches("pay-unless/inti-seneschal-of-the-sun-discard",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenYouDo("Inti, Seneschal of the Sun — put a +1/+1 counter on target attacking creature; it gains trample until end of turn",
			intiCounterBody).Apply(ctx)
	}, nil)

// intiCounterBody is the reflexive trigger's registration: target
// attacking creature gets a +1/+1 counter and trample until end of turn.
var intiCounterBody = game.ReflexiveBody("inti-seneschal-of-the-sun/counter-and-trample",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.UntilEndOfTurn(ctx, target, nil, "Inti, Seneschal of the Sun — trample", game.AddKeywordsMod("trample"))
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target attacking creature", effects.AttackingCreature())
	}))
