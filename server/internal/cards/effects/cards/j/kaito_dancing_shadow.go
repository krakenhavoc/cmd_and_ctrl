package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaito, Dancing Shadow — Legendary Planeswalker — Kaito {2}{U}{B},
// starting loyalty 3 (Phyrexia: All Will Be One, #5479):
//
//	"Whenever one or more creatures you control deal combat damage to a
//	 player, you may return one of them to its owner's hand. If you do, you
//	 may activate loyalty abilities of Kaito twice this turn rather than
//	 only once.
//	 +1: Up to one target creature can't attack or block until your next
//	     turn.
//	 0: Draw a card.
//	 −2: Create a 2/2 colorless Drone artifact creature token with
//	     deathtouch and "When this token leaves the battlefield, each
//	     opponent loses 2 life and you gain 2 life.""
//
// The trigger is one per player the creatures connect with (CR 603.2c);
// "one of them" is a creature of that batch still on the battlefield, and
// returning it is optional (an empty pick declines). "If you do" is the
// bounce reaching a hand, and then the permission is Kaito's alone:
// game.AllowThisPermanentLoyaltyTwiceThisTurnForEffect, a one-turn effect
// that an activation already made this turn counts toward, so it may be
// used after the first activation. A Kaito that has left the battlefield
// gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17fd42f6-f388-4ffc-8fec-ca9ada2506ee",
		Name:            "Kaito, Dancing Shadow",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Triggered: []game.TriggeredAbility{
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Kaito, Dancing Shadow — you may return one of them to its owner's hand; if you do, you may activate loyalty abilities of Kaito twice this turn",
				kaitoDancingShadowOffer),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature can't attack or block until your next turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  kaitoDancingShadowPlusOne,
			},
			{
				Label:   "0: Draw a card.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label: "−2: Create a 2/2 colorless Drone artifact creature token with deathtouch and \"When this token leaves the battlefield, each opponent loses 2 life and you gain 2 life.\"",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.S82OneB20DroneToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

func kaitoDancingShadowPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   targets[0].ID,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)},
		Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Kaito, Dancing Shadow — can't attack or block until your next turn",
	}.Apply(ctx)
}

// kaitoDancingShadowOffer asks which of the creatures that connected to
// return (nothing is a legal answer).
func kaitoDancingShadowOffer(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	var candidates []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, prior := range g.Events {
		if prior.Batch != ev.Batch || prior.Kind != game.EventDealDamage || !prior.Combat ||
			prior.Amount <= 0 || prior.Target != ev.Target || seen[prior.Source] {
			continue
		}
		seen[prior.Source] = true
		c, ok := g.LookupCardForEffect(prior.Source)
		if ok && c.IsCreature() && c.Controller == item.Controller && g.Battlefield.Contains(prior.Source) {
			candidates = append(candidates, prior.Source)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     ctx.Source(),
		Question:   "Kaito, Dancing Shadow — you may return one of the creatures that dealt combat damage to its owner's hand",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    kaitoDancingShadowReturnPicked,
	})
	return nil
}

// kaitoDancingShadowReturnPicked bounces the pick. An on-disk key: never
// renamed, never reused.
var kaitoDancingShadowReturnPicked = game.RegisterChooseCardsThen("choose-cards/kaito-dancing-shadow-return",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		return g.BounceCardsToHandKeyedForEffect(r.Picked, game.KeyedTail{
			Ref: kaitoDancingShadowBounced, Player: r.Chooser, Source: r.Source,
		})
	})

// kaitoDancingShadowBounced is "if you do": the creature reached a hand,
// so Kaito's loyalty abilities may be activated twice this turn. An
// on-disk key.
var kaitoDancingShadowBounced = effects.CardTail("then/kaito-dancing-shadow-bounced", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	ctx.Game.AllowThisPermanentLoyaltyTwiceThisTurnForEffect(o.Tail.Source)
	return nil
})
