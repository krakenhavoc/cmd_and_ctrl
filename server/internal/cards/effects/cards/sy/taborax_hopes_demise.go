package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taborax, Hope's Demise — Legendary Creature — Demon Cleric {2}{B}, 2/2
// (Foundations Jumpstart, #5929):
//
//	"Flying
//	 Taborax has lifelink as long as it has five or more +1/+1 counters on it.
//	 Whenever another nontoken creature you control dies, put a +1/+1 counter
//	 on Taborax. If that creature was a Cleric, you may draw a card. If you
//	 do, you lose 1 life."
//
// The lifelink is Primordial Hydra's layer 6 self static on the live
// counter count. The trigger is Body Launderer's: another nontoken creature
// its controller controlled as it died (read from the event, CR 603.10a). The
// counter goes on first; the Cleric rider reads the dead creature's types as
// it last existed (a changeling was one; a creature that was only a Cleric
// because of an effect that ended still counts), and its draw is a real
// "you may": declining, or a draw a replacement removes, costs no life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "955abada-aba2-41c2-96ff-58e3679ad7e5",
		Name:            "Taborax, Hope's Demise",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label: "Taborax has lifelink as long as it has five or more +1/+1 counters on it.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.Counters[game.CounterPlusOne] >= 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "lifelink")
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B19AnotherNontokenCreatureYouControlDied(ev, source, g)
			}, "Taborax — put a +1/+1 counter on it; if that creature was a Cleric, you may draw a card, and if you do, you lose 1 life",
				taboraxCreatureDied),
		},
	})
}

func taboraxCreatureDied(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	ev := item.Trigger.Event
	dead, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !effects.LeftAsSubtype(ev, dead, ctx.TextSubtype("Cleric")) {
		return nil
	}
	return effects.MayChoice{
		Question: "Taborax, Hope's Demise — draw a card? If you do, you lose 1 life",
		Draws:    1,
		YesLabel: "Draw",
		NoLabel:  "Decline",
		Then:     taboraxDrawAnswer,
	}.Apply(ctx)
}

// taboraxDrawAnswer is the "yes": draw a card, then lose 1 life. Keyed.
var taboraxDrawAnswer = effects.MayYes("confirm/bns-b52-taborax-draw-lose-life", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1})(ctx.Game, ctx.Item)
})
