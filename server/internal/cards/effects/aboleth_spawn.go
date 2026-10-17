package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aboleth Spawn — Creature — Fish Horror {2}{U}, 2/3 (#4092):
//
//	"Flash
//	 Ward {2}
//	 Probing Telepathy — Whenever a creature entering under an
//	 opponent's control causes a triggered ability of that creature to
//	 trigger, you may copy that ability. You may choose new targets for
//	 the copy."
//
// Probing Telepathy watches EventTriggerPlaced (#4092, #4050): the
// drain announces every triggered ability it puts on the stack, after
// the whole batch is there, so the Spawn's own trigger is the next
// batch and lands ABOVE the ability it copies whichever seat is active.
// That is the 2022-06-10 ruling word for word: "Aboleth Spawn's
// ability will always go on the stack on top of the entering
// creature's triggered ability that caused it to trigger", and the
// copy resolves before the original.
//
// The filter reads the placed item's own triggering event
// (StackItem.Trigger, #1223) rather than the board:
//
//   - the event is an entry (isEntering, the doublers' reading) and the
//     object that entered is the ability's SOURCE — "a triggered
//     ability of that creature", so a Soul Warden that sees the
//     creature enter is not copied (ruling: "cares only about triggered
//     abilities of the creature that's entering");
//   - the entering object was a creature, and under an opponent's
//     control, as it entered: the trigger's ObjectSnapshot, taken at
//     harvest time, not the permanent now.
//
// "You may copy" is the resolution-time CR 603.5 prompt (Optional), and
// the copy is CopyAbility with CR 707.10c's re-target offer, the Rings
// of Brighthearth path. The ability to copy is named by the triggering
// event's stack item id. If it has left the stack by then (Stifled in
// response) nothing is copied, which is the weaker reading.
//
// Entering as a copy, with counters or tapped is a replacement, not a
// triggered ability (CR 614.1c), so it never reaches this event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6ae77b3-59e7-475c-800c-930a5cb2e0ca",
		Name:            "Aboleth Spawn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Aboleth Spawn — ward {2}"),
			Optional(
				On(game.EventTriggerPlaced, abolethSpawnEntryTrigger,
					"Aboleth Spawn — Probing Telepathy: copy that ability", copyTheTriggeringAbility),
				"Aboleth Spawn — copy that ability?"),
		},
	})
}

// abolethSpawnEntryTrigger is Probing Telepathy's trigger condition: the
// triggered ability just put on the stack belongs to a creature whose
// entering under an opponent's control caused it to trigger.
func abolethSpawnEntryTrigger(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.StackItemID == uuid.Nil {
		return false
	}
	item := g.StackItemForEffect(ev.StackItemID)
	if item == nil || item.Kind != game.StackItemTriggered || item.Trigger == nil {
		return false
	}
	cause := item.Trigger.Event
	if !isEntering(cause) || cause.CardID == uuid.Nil || cause.CardID != item.SourceCardID {
		return false
	}
	entered := item.Trigger.Object
	if entered == nil || !entered.HasType("Creature") {
		return false
	}
	return entered.Controller != uuid.Nil && entered.Controller != source.Controller
}
