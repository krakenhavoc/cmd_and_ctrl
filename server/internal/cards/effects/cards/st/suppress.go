package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suppress — Sorcery {2}{B} (Apocalypse, #5614):
//
//	"Target player exiles all cards from their hand face down. At the
//	 beginning of the end step of that player's next turn, that player returns
//	 those cards to their hand."
//
// The hand is exiled face down (Memory Jar's exile: CR 406.3, nobody may look
// at them), in one batch. Once the last card has landed, the chain's tail
// schedules the CR 603.7 delayed trigger with every card exiled this way and
// the CR 400.7 epoch each landed with, as plain data: it fires at the
// beginning of the end step of that player's next turn, never the turn it was
// scheduled in (so a player who targets themselves waits for their following
// turn). When it resolves, each card still in exile as the same object returns
// to its owner's hand, who is the player who exiled it (CR 400.3); a card that
// left exile in between is a new object and stays where it is. An empty hand
// exiles nothing and schedules nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b39651ad-1c05-45f8-b036-536c4d7324c0",
		Name:         "Suppress",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.ExileHandFaceDownKeyedForEffect(player, ctx.ItemTail(suppressSchedule))
		},
	})
}

// suppressSchedule runs once the hand is exiled, told every card that landed in
// exile: it schedules the return at the end step of the player's next turn. An
// on-disk key: never renamed, never reused.
var suppressSchedule = effects.CardTail("s82-old-b635/suppress-schedule", func(ctx *effects.Context, o game.TailOutcome) error {
	player := effects.TargetedPlayerOf(ctx)
	var objs []game.ObjectRef
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			objs = append(objs, game.ObjectRef{ID: id, Epoch: ctx.Game.ObjectEpochForEffect(id)})
		}
	}
	if len(objs) == 0 || player == uuid.Nil {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:            game.StepEnd,
		Label:         "Suppress — that player returns the exiled cards to their hand",
		TurnOf:        player,
		AfterThisTurn: true,
		Body:          suppressReturn,
		Params:        game.EffectParams{Objects: objs},
	}.Apply(ctx)
})

// suppressReturn is the delayed trigger: each card still the exiled object
// goes to its owner's hand. An on-disk key: never renamed, never reused.
var suppressReturn = game.DelayedBody("s82-old-b635/suppress-return", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	var back []uuid.UUID
	for _, ref := range item.Params.Objects {
		z := g.FindCardZoneForEffect(ref.ID)
		if z != nil && z.Kind == game.ZoneExile && g.ObjectEpochForEffect(ref.ID) == ref.Epoch {
			back = append(back, ref.ID)
		}
	}
	if len(back) > 0 {
		g.BounceCardsToHandForEffect(back)
	}
	return nil
})
