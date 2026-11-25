package game

import "github.com/google/uuid"

// random_events.go — "whenever you discard one or more cards at random,
// select one or more targets at random, choose one or more modes at
// random, …" (Chira, All In; #6101). Coins and dice already announce
// themselves (EventFlipCoin, EventRollDie, one per coin or die with a
// shared BatchSeq). The three other random choices each get one event per
// instruction here, emitted where the engine makes the draw, so a card
// that watches them never has to know which card or cost asked for it:
//
//   - EventDiscardedAtRandom: a player discards one or more cards at
//     random (CR 701.9b), by an effect (DiscardRandomKeyedForEffect, which
//     every catalog "discards … at random" reaches), by an activated
//     ability's "discard N cards at random" cost, or by a pay-unless
//     "discard N cards at random" payment. Actor is the discarding
//     player, Amount the number of cards drawn for the discard.
//   - EventTargetsSelectedAtRandom: a target drawn at random as an
//     ability is put on the stack (TargetSpec.AtRandom: the trigger target
//     walk and an activation's announcement), or a spell's or ability's
//     targets reselected at random (ReselectTargetsAtRandomForEffect)
//     when at least one slot had a target to draw. Actor is the player
//     whose ability or effect made the selection.
//   - EventModesChosenAtRandom: "choose one at random" (CR 700.2,
//     ModeSpec.AtRandom) picked a mode. Actor is the ability's
//     controller.
//
// CardID is the source whose ability or effect made the choice, when
// there is one. The events are announcements, not history the engine
// replays: nothing reads them but triggered abilities and the log.

const (
	EventDiscardedAtRandom       EventKind = "discarded_at_random"
	EventTargetsSelectedAtRandom EventKind = "targets_selected_at_random"
	EventModesChosenAtRandom     EventKind = "modes_chosen_at_random"
)

// emitChoseAtRandomLocked emits one of the three events above for one
// instruction.
//
// Caller must hold g.mu.
func (g *Game) emitChoseAtRandomLocked(kind EventKind, actor, source uuid.UUID, n int) {
	if actor == uuid.Nil || n <= 0 {
		return
	}
	g.EmitEvent(Event{Kind: kind, Actor: actor, CardID: source, Amount: n})
}
