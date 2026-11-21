package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ugin, the Ineffable — Legendary Planeswalker — Ugin {6}, loyalty 4 (Modern
// Horizons 3 Commander, #5948):
//
//	"Colorless spells you cast cost {2} less to cast.
//	 +1: Exile the top card of your library face down and look at it.
//	 Create a 2/2 colorless Spirit creature token. When that token leaves the
//	 battlefield, put the exiled card into your hand.
//	 −3: Destroy target permanent that's one or more colors."
//
// The discount is Eye of Ugin's, generic-only and read off the spell being
// cast (a spell with any colour is not discounted). The +1 exiles the top card
// face down and linked to Ugin (ExileHiddenLinkedForEffect: its controller may
// look at it, nobody else may), then creates the Spirit whether or not a card
// was exiled. "That token" is the one the creating effect made (a token tail is
// told the new object): the leave trigger is a CR 603.7 delayed trigger on that
// object, with no duration but its own firing (Mechtitan Core's shape), fired
// once by the token's EventLTB however it left. Its body puts the card into its
// owner's hand only while it is still the same face-down card in exile (CR
// 400.7). The −3 targets a permanent with at least one colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "18bdc181-9592-4147-81fb-7f83ce137f70",
		Name:            "Ugin, the Ineffable",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Colorless spells you cast cost {2} less to cast.",
				effects.YourSpell(), func(q game.CostQuery) bool { return q.Card.IsColorless() }),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Exile the top card of your library face down and look at it. Create a 2/2 colorless Spirit creature token. When that token leaves the battlefield, put the exiled card into your hand.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  uginTheIneffablePlusOne,
			},
			{
				Label:   "−3: Destroy target permanent that's one or more colors.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target permanent that's one or more colors", effects.Not(effects.Colorless())),
				Effect:  effects.DestroyFirstLegalCardTarget,
			},
		},
	})
}

func uginTheIneffablePlusOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var exiled []uuid.UUID
	if looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1); len(looked) > 0 {
		if _, err := g.ExileHiddenLinkedForEffect(effects.HiddenRefOfActivation(item), item.Controller, looked[0]); err != nil {
			return err
		}
		exiled = looked
	}
	tail := ctx.Tail(uginTheIneffableSpiritMade)
	tail.Carry = exiled
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(effects.TokenCard("2/2 colorless Spirit")),
			Count:    1,
		}},
	}, tail)
}

var (
	// "When that token leaves the battlefield": the one object CondParams.
	// Object names, whichever way it left.
	uginTheIneffableLeftCondition = game.DelayedEachCondition("ltb/ugin-the-ineffable-spirit", effects.MechtitanLeft)

	// "Put the exiled card into your hand." The card rides the trigger's Cards.
	// An on-disk key.
	uginTheIneffableToHandBody = game.SimpleDelayedBody("ugin-the-ineffable/exiled-card-to-hand", effects.NecropotenceDeliverExiled)
)

// uginTheIneffableSpiritMade schedules the leave trigger for the Spirit just
// made, carrying the card the +1 exiled.
var uginTheIneffableSpiritMade = effects.CardTail("ugin-the-ineffable/spirit-made", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	token := o.IDs[0]
	c, ok := effects.BattlefieldCardForEffect(ctx.Game, token)
	if !ok {
		return nil
	}
	end := game.IndefiniteDuration()
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        "Ugin, the Ineffable — put the exiled card into your hand",
		On:           []game.EventKind{game.EventLTB},
		Condition:    uginTheIneffableLeftCondition,
		CondParams:   game.EffectParams{Object: game.ObjectRef{ID: token, Epoch: c.ObjectEpoch}},
		Cards:        append([]uuid.UUID(nil), o.Tail.Carry...),
		Body:         uginTheIneffableToHandBody,
		Duration:     &end,
	})
	return nil
})
