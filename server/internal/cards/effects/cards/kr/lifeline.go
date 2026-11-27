package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lifeline — Artifact {5} (Urza's Saga, S82 OLD batch 730, #5624):
//
//	"Whenever a creature dies, if another creature is on the battlefield,
//	 return the first card to the battlefield under its owner's control at the
//	 beginning of the next end step."
//
// "The first card" is the creature that died. The intervening "if" (CR 603.4)
// is another creature on the battlefield, anyone's, through the layers: it is
// checked as the creature dies and again as the trigger resolves, so two
// creatures dying together with none left trigger nothing, and the last one
// leaving in response does nothing. Resolving schedules a CR 603.7 delayed
// trigger for the next end step that returns the card under its owner's
// control if it is still in the graveyard as the object that died (CR 400.7);
// that return does not look at the board again. A token never reaches a
// graveyard and is not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f475b636-eb7f-4dd7-bc87-a1af07064ffe",
		Name:         "Lifeline",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "another creature is on the battlefield",
				Test: func(ctx *effects.Context) bool { return effects.OldB730AnotherCreatureIsOnTheBattlefield(ctx.Game) },
			}, effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && effects.OldB730AnotherCreatureIsOnTheBattlefield(g)
			}, "Lifeline — return the first card to the battlefield at the beginning of the next end step", lifelineSchedule)),
		},
	})
}

// lifelineSchedule schedules the end-step return of the creature card the
// trigger is about, if it is still in a graveyard as the card that died.
func lifelineSchedule(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	id := obj.Ref().ID
	if !effects.DiedCardStillInGraveyard(ctx, id) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Lifeline — return the card to the battlefield under its owner's control",
		Cards:  []uuid.UUID{id},
		Body:   effects.OldB730ReturnTheCardToOwnerBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}
