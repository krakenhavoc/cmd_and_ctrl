package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Back for More — Instant {4}{B}{G}:
//
//	"Return target creature card from your graveyard to the battlefield.
//	 When you do, it fights up to one target creature you don't control."
//
// The second sentence is a reflexive trigger (CR 603.12) created only
// when the creature really returned. Its target is chosen as it goes on
// the stack and may be none ("up to one"); the returned creature rides
// the payload and fights the chosen creature only while both are still
// on the battlefield (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "670df8be-21cd-49d8-92e3-d7cc38b256d6",
		Name:         "Back for More",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			card, ok := effects.ReanimateSingleTarget(ctx, ctx.Controller())
			if !ok {
				return nil
			}
			if z := ctx.Game.FindCardZoneForEffect(card.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
				return nil
			}
			return effects.ReflexiveTrigger{
				Label: "Back for More — the returned creature fights up to one target creature you don't control",
				Cards: []uuid.UUID{card.InstanceID},
				Body:  backForMoreFightBody,
			}.Apply(ctx)
		},
	})
}

var backForMoreFightBody = game.ReflexiveBody("s82-bns-b02/back-for-more-fight",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		returned := ctx.PayloadCards()
		if len(returned) == 0 {
			return nil
		}
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.B10Fight(ctx, returned[0], id)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1)
	}))
