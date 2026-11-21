package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mechtitan Core — Artifact — Vehicle {2}, 2/4 (Neon Dynasty):
//
//	"{5}, Exile this Vehicle and four other artifact creatures and/or
//	 Vehicles you control: Create Mechtitan, a legendary 10/10 Construct
//	 artifact creature token with flying, vigilance, trample, lifelink, and
//	 haste that's all colors. When that token leaves the battlefield, return
//	 all cards exiled with this Vehicle except this card to the battlefield
//	 tapped under their owners' control.
//	 Crew 2"
//
// The cost is {5}, ExileThis and four OTHER materials with the clause
// "artifact creature and/or Vehicle" (ExilePermanentsCost.
// ArtifactCreatureOrVehicle): the board's own permanents, chosen at
// announce, tokens included. What the four exile is PaidCost.Exiled, which
// never holds the Core itself ("except this card"), so that record is the
// "cards exiled with this Vehicle" set (a token among them has ceased to
// exist and cannot return, CR 111.8).
//
// "That token" is the one the creating effect made (a token tail is told the
// new object). The leave trigger is a CR 603.7 delayed trigger on that
// object, with no duration but its own firing (the Xira shape), fired once by
// the token's EventLTB however it left and read as the object it was (CR
// 400.7). Its body returns the exiled cards together, tapped, under their
// owners' control (CR 603.6a's batch entry; ReturnFromExileTogether).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1eb489d-104c-4801-b6d2-3f1a7ee73a75",
		Name:         "Mechtitan Core",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Crew 2",
				Cost:    effects.CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Mechtitan Core"),
			},
			{
				Label:   "{5}, Exile this Vehicle and four other artifact creatures and/or Vehicles you control: Create Mechtitan",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive},
				Cost: effects.Plus(effects.ManaCost("{5}"), effects.ExileThis(), game.AbilityCost{ExilePermanents: &game.ExilePermanentsCost{
					Count:                     4,
					ExcludeSource:             true,
					ArtifactCreatureOrVehicle: true,
					Label:                     "four other artifact creatures and/or Vehicles you control",
				}}),
				Effect: mechtitanCoreCreate,
			},
		},
	})
}

const mechtitanToken = "10/10 white and blue and black and red and green Legendary Construct artifact named Mechtitan with flying, vigilance, trample, lifelink, and haste"

func mechtitanCoreCreate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(effects.TokenCard(mechtitanToken)),
			Count:    1,
		}},
	}, mechtitanTail(ctx))
}

var (
	// "When that token leaves the battlefield": the one object CondParams.
	// Object names, whichever way it left.
	mechtitanLeftCondition = game.DelayedEachCondition("ltb/mechtitan-token", mechtitanLeft)

	// "Return all cards exiled with this Vehicle except this card to the
	// battlefield tapped under their owners' control." The cards ride on the
	// trigger's Cards. An on-disk key.
	mechtitanReturnBody = game.SimpleDelayedBody("mechtitan-core/return-exiled", mechtitanReturn)
)

func mechtitanLeft(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if p.Object.ID == uuid.Nil || ev.Kind != game.EventLTB || ev.CardID != p.Object.ID {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(ev.CardID)
	if !ok || ref.Epoch != p.Object.Epoch {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

// mechtitanTail is the creation's continuation with the cards the cost
// exiled carried as data: a tail's item copy does not keep the payment.
func mechtitanTail(ctx *effects.Context) game.KeyedTail {
	t := ctx.Tail(mechtitanCoreCreated)
	t.Carry = append([]uuid.UUID(nil), ctx.Paid().Exiled...)
	return t
}

// mechtitanCoreCreated schedules the leave trigger for the token just made,
// carrying the cards the cost exiled.
var mechtitanCoreCreated = effects.CardTail("mechtitan-core/token-created", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
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
		Label:        "Mechtitan Core — return all cards exiled with this Vehicle to the battlefield tapped",
		On:           []game.EventKind{game.EventLTB},
		Condition:    mechtitanLeftCondition,
		CondParams:   game.EffectParams{Object: game.ObjectRef{ID: token, Epoch: c.ObjectEpoch}},
		Cards:        append([]uuid.UUID(nil), o.Tail.Carry...),
		Body:         mechtitanReturnBody,
		Duration:     &end,
	})
	return nil
})

func mechtitanReturn(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	return effects.ReturnFromExileTogether{Targets: ids, Tapped: true}.Apply(effects.NewContext(g, item))
}
