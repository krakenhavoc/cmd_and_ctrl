package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strict Proctor — Creature — Spirit Cleric {1}{W}, 1/3 (#6158):
//
//	"Flying
//	 Whenever a permanent entering causes a triggered ability to
//	 trigger, counter that ability unless its controller pays {2}."
//
// The trigger watches EventTriggerPlaced, Aboleth Spawn's seam (#4092):
// the CR 603.3b drain announces every triggered ability it puts on the
// stack after the whole batch is there, so the Proctor's ability is the
// next batch and goes on the stack ABOVE the ability it counters,
// whichever seat controls either. That is CR 603.3b's second part: an
// ability whose trigger condition is another ability triggering is put
// on the stack after the ones whose condition is not.
//
// The filter reads the placed item's own triggering event
// (StackItem.Trigger): the ability triggered on a permanent entering
// the battlefield (IsEntering, the doublers' reading), any permanent and
// any triggered ability, the entering permanent's own or another
// permanent's that saw it enter. A copy of an ability is created, not
// triggered, and is never announced here (CR 707.10a).
//
// "Counter that ability unless its controller pays {2}" is
// CounterUnlessPaid on the placed item: the ability's controller is
// asked as the Proctor's ability resolves (CR 118.12), and an ability
// that has already left the stack raises no prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b967870d-9773-4ad5-bf06-c3be5465dab9",
		Name:            "Strict Proctor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTriggerPlaced, strictProctorEntryCausedTrigger,
				"Strict Proctor — counter that ability unless its controller pays {2}", strictProctorCounter),
		},
	})
}

// strictProctorEntryCausedTrigger is the trigger condition: the
// triggered ability just put on the stack triggered on a permanent
// entering the battlefield.
func strictProctorEntryCausedTrigger(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.StackItemID == uuid.Nil {
		return false
	}
	item := g.StackItemForEffect(ev.StackItemID)
	if item == nil || item.Kind != game.StackItemTriggered || item.Trigger == nil {
		return false
	}
	return effects.IsEntering(item.Trigger.Event)
}

func strictProctorCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ability := ctx.Trigger().Event.StackItemID
	if ability == uuid.Nil {
		return nil
	}
	return effects.CounterUnlessPaid{
		StackID:  ability,
		Cost:     "{2}",
		Question: "Strict Proctor — pay {2} or your triggered ability is countered",
	}.Apply(ctx)
}
