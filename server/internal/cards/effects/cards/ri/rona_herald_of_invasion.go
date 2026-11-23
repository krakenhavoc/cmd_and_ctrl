package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rona, Herald of Invasion // Rona, Tolarian Obliterator — {1}{U}
// Legendary Creature — Human Wizard 1/3 // Legendary Creature — Phyrexian
// Wizard 5/5 (March of the Machine):
//
//	Front: "Whenever you cast a legendary spell, untap Rona.
//	        {T}: Draw a card, then discard a card.
//	        {5}{B/P}: Transform Rona. Activate only as a sorcery."
//	Back:  "Trample
//	        Whenever a source deals damage to Rona, that source's controller
//	        exiles a card from their hand at random. If it's a land card,
//	        you may put it onto the battlefield under your control.
//	        Otherwise, you may cast it without paying its mana cost."
//
// One trigger per damage event, the dealer's controller captured as it
// fires (b17DamageSourceController). On resolution that player's hand is
// drawn from at random through the game's keyed random stream (CR 701.9b:
// the hand as it is then) and the card is exiled. "You" is Rona's
// controller: a land may be put onto the battlefield under their control
// (declined, it stays in exile), anything else may be cast by them
// without paying its mana cost (declined, it stays in exile).
//
// No simplification.
func init() {
	const oracle = "cf5bd3a4-3899-4c25-ad1f-f1778c1d9c60"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Rona, Herald of Invasion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Legendary(), "Rona, Herald of Invasion — untap Rona", effects.B31UntapSelf),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.LootOne(g, item, 1)
				},
			},
			{
				Label:        "{5}{B/P}: Transform Rona. Activate only as a sorcery.",
				Cost:         effects.ManaCost("{5}{B/P}"),
				SorcerySpeed: true,
				Effect:       effects.Do(effects.TransformThis{}),
			},
		},
	})
	const label = "Rona, Tolarian Obliterator — that source's controller exiles a card from their hand at random; you may put it onto the battlefield (land) or cast it free"
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Rona, Tolarian Obliterator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B17SourceDealtDamageToSelf(ev, source)
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = effects.B17DamageSourceController(ev, g)
				return item
			},
			Effect: ronaObliteratorExile,
		}},
	})
}

// ronaObliteratorExile exiles a random card from the damaging source's
// controller's hand.
func ronaObliteratorExile(g *game.Game, item *game.StackItem) error {
	victim := g.PlayerByIDForEffect(item.Params.Player)
	if victim == nil || victim.Eliminated || victim.Hand == nil || len(victim.Hand.Cards) == 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	hand := make([]uuid.UUID, 0, len(victim.Hand.Cards))
	for _, c := range victim.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	picked := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), hand, 1)
	if len(picked) == 0 {
		return nil
	}
	return effects.ExileTarget{Target: picked[0], ThenTail: ctx.ItemTail(ronaObliteratorExiled)}.Apply(ctx)
}

// ronaObliteratorExiled is the rest of the sentence, told the card that
// reached exile: a land may go onto the battlefield under Rona's
// controller's control, anything else may be cast free by them. An
// on-disk key: never renamed, never reused.
var ronaObliteratorExiled = effects.CardTail("rona-tolarian-obliterator/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	id := o.IDs[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if !c.IsLand() {
		return ctx.Game.CastExiledFreeKeepingForEffect(ctx.Controller(), ctx.Source(), id)
	}
	return effects.MayChoice{
		Question: "Rona, Tolarian Obliterator — put " + c.Name + " onto the battlefield under your control?",
		YesLabel: "Put it onto the battlefield",
		NoLabel:  "Leave it in exile",
		Then:     ronaObliteratorLandAnswer,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
})

var ronaObliteratorLandAnswer = effects.MayAnswer("confirm/rona-tolarian-obliterator-land", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if !r.Accepted || len(r.Carry) == 0 {
		return nil
	}
	return effects.ReturnFromExile{Target: r.Carry[0], Controller: ctx.Controller()}.Apply(ctx)
})
