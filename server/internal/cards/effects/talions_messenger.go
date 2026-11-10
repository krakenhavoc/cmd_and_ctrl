package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Talion's Messenger — Creature — Faerie Noble {2}{U}, 3/2 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Whenever you attack with one or more Faeries, draw a card, then
//	 discard a card. When you discard a card this way, put a +1/+1 counter
//	 on target Faerie you control."
//
// The attack trigger is Hired Claw's "attack with one or more Lizards" over
// Faeries (one trigger per attack declaration whatever the number, a word of
// the Messenger's text, CR 612.2). The loot is drawThen's "draw, then
// discard", the discard mandatory; once a card was really discarded
// (an empty hand discards nothing) the reflexive trigger of CR 603.12 is
// created, and picks its Faerie as it goes on the stack, so the discard is
// answered before the target is chosen (the ruling), as Tether Technician's
// is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a721a4c-696b-44d1-858d-20040bd01909",
		Name:            "Talion's Messenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && attackerHasSubtype(g, ev, source.TextSubtype("Faerie"))
			}, "Talion's Messenger — draw a card, then discard a card; when you do, put a +1/+1 counter on target Faerie you control",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: talionsMessengerLoot, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				})),
		},
	})
}

// talionsMessengerLoot is the discard after the draw: one card, not
// optional, with the reflexive counter keyed to it. An on-disk key.
var talionsMessengerLoot = game.RegisterDrawThen("talions-messenger/discard", func(g *game.Game, d game.DrawThen) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   d.Player,
		Source:   d.Source,
		N:        d.N,
		Question: "Talion's Messenger — discard a card",
	}, talionsMessengerDiscarded, nil)
})

var talionsMessengerDiscarded = game.RegisterDiscardThen("choose-cards/talions-messenger-counter", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Discarded.Cards()) == 0 {
		return nil
	}
	return WhenYouDo("Talion's Messenger — put a +1/+1 counter on target Faerie you control",
		talionsMessengerCounterBody).Apply(keyedContext(g, r.Player, r.Source))
})

var talionsMessengerCounterBody = game.ReflexiveBody("talions-messenger/counter",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target Faerie you control", YouControl(), OfSubtype("Faerie"))
	}))
