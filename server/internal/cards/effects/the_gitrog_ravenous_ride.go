package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Gitrog, Ravenous Ride — Legendary Creature — Frog Horror Mount
// {3}{B}{G}, 6/5:
//
//	"Trample, haste
//	 Whenever The Gitrog deals combat damage to a player, you may
//	 sacrifice a creature that saddled it this turn. If you do, draw X
//	 cards, then put up to X land cards from your hand onto the
//	 battlefield tapped, where X is the sacrificed creature's power.
//	 Saddle 1"
//
// Trample and haste are PrintedKeywords; Saddle 1 is Saddle(1).
//
// The trigger's candidates are SaddlersOfThis filtered to creatures its
// controller controls (ADR 0186 §1): read by The Gitrog's object, so a
// Gitrog that died to the very combat damage that triggered it still
// offers its saddlers. "You may sacrifice" is a card-set prompt with a
// floor of zero; declining sacrifices nothing and draws nothing. The
// sacrifice goes through SacrificePermanent, whose continuation says
// whether the creature really left, and X is its power as it last existed
// on the battlefield (CR 608.2h, the read Ghoulcaller Gisa makes), so a
// pumped creature counts its pump. The put is one entry of up to X lands,
// tapped, after the draw (a land the draw finds can be put).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7739311a-682b-48a6-8db2-c07f44141751",
		Name:            "The Gitrog, Ravenous Ride",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated:       []ActivatedAbility{Saddle(1)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Key:    gitrogRavenousRideLabel,
			Effect: gitrogRavenousRideSacrifice,
		}},
	})
}

const gitrogRavenousRideLabel = "The Gitrog, Ravenous Ride — you may sacrifice a creature that saddled it, then draw and put lands"

// gitrogRavenousRideSacrifice is the combat-damage trigger's body.
func gitrogRavenousRideSacrifice(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, id := range SaddlersOfThis(NewContext(g, item)) {
		if c, ok := g.LookupCardForEffect(id); ok && c.Controller == item.Controller {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "The Gitrog, Ravenous Ride — you may sacrifice a creature that saddled it",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  gitrogRavenousRidePick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// gitrogRavenousRidePick is the saddling creature chosen to be sacrificed,
// keyed (#4253 sweep): the asking item rides the prompt as data, and the
// sacrifice's rest is a keyed tail.
var gitrogRavenousRidePick = askedPick("choose-cards/the-gitrog-ravenous-ride-sacrifice", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return SacrificePermanent{Target: r.Picked[0], ThenTail: withCarry(ctx.ItemTail(gitrogRavenousRideSacrificed), r.Picked[0])}.Apply(ctx)
})

// gitrogRavenousRideSacrificed is "draw X cards, then put up to X lands",
// X being the power the sacrificed creature last had, keyed (#4253): the
// creature rides the tail as its carried ID, and the tail is told whether
// it went. An on-disk key: never renamed, never reused.
var gitrogRavenousRideSacrificed = CardTail("the-gitrog-ravenous-ride/sacrificed", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	info, ok := ctx.Game.LastKnownPermanentForEffect(tailCarry0(o))
	if !ok || info.Power <= 0 {
		return nil
	}
	ctx.RecordStep(gitrogRavenousRideRecord, nil, info.Power)
	return gitrogRavenousRideDrawAndPut.Run(ctx)
})

// gitrogRavenousRideRecord is X, the sacrificed creature's power.
const gitrogRavenousRideRecord = "the-gitrog-ravenous-ride/x"

// gitrogRavenousRideDrawAndPut is "draw X cards, then put up to X land cards
// from your hand onto the battlefield tapped": the put waits for a prompt the
// draw asks (a dredge offer; CR 608.2c, #4155).
var gitrogRavenousRideDrawAndPut = RegisterThenSteps("then-steps/the-gitrog-ravenous-ride-draw",
	Instruction(func(ctx *Context) error {
		return DrawCards{Player: ctx.Controller(), N: ctx.StepRecord(gitrogRavenousRideRecord).N()}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		return PutFromHandOntoBattlefield{
			Player:   ctx.Controller(),
			Match:    Land(),
			Tapped:   true,
			Optional: true,
			Max:      ctx.StepRecord(gitrogRavenousRideRecord).N(),
			Label:    "The Gitrog, Ravenous Ride — put up to X land cards from your hand onto the battlefield tapped",
		}.Apply(ctx)
	}),
)
