package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grip of Chaos — Enchantment {4}{R}{R} (#4050):
//
//	"Whenever a spell or ability is put onto the stack, if it has a
//	 single target, reselect its target at random. (Select from among
//	 all legal targets.)"
//
// "Put onto the stack" is every way an object gets there, and each is
// its own event:
//
//   - a spell is cast (EventCast, CardID is the spell) or copied
//     (EventSpellCopied, CardID is the copy; CR 707.10);
//   - an activated ability is activated (EventActivateAbility,
//     StackItemID), loyalty abilities included; a mana ability never
//     uses the stack (CR 605.3a) and is not watched;
//   - a triggered ability is placed by the CR 603.3b drain
//     (EventTriggerPlaced, #4099);
//   - a copy of an activated or triggered ability is created
//     (EventAbilityCopied), which fires neither of the two above
//     (CR 707.10a).
//
// "If it has a single target" is CR 115.9a, counted by instances of the
// word "target" (StackItemTargetCountForEffect counts slots, so a spell
// aimed at one creature twice has two targets and is left alone). It is
// an intervening if (CR 603.4): checked when the event happens, and
// again as the trigger resolves, where the item may have left the stack
// or changed shape in response. The target is then chosen at
// resolution, "at which time the set of legal targets may have changed"
// (2004-10-04 ruling), by ReselectTargetsAtRandomForEffect (#1815):
// every legal target for the item's controller, the one it has now
// included ("it may still choose the original one"), drawn from the
// keyed random stream so an undo replays the same pick and nobody is
// asked.
//
// The trigger goes on the stack above the object it watched, so it
// resolves first and the spell or ability then goes on to resolve with
// its new target. It cannot loop: its own trigger targets nothing, so
// the intervening if fails for it (as it does for a second Grip's
// trigger watching the first), and the reselection rewrites an item in
// place rather than putting anything onto the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca2e0c8f-7ece-465f-9103-ba60712eacf7",
		Name:         "Grip of Chaos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{
				game.EventCast, game.EventSpellCopied,
				game.EventActivateAbility, game.EventTriggerPlaced, game.EventAbilityCopied,
			}, putOntoTheStackWithASingleTarget,
				"Grip of Chaos — reselect its target at random", gripReselectsTheTarget),
		},
	})
}

// putOntoTheStackItem is the stack item id an event of any of Grip of
// Chaos' five kinds is about: the spell's card for a cast or a spell
// copy (a spell's item id is its card's instance id), the ability item
// for the other three.
func putOntoTheStackItem(ev game.Event) uuid.UUID {
	switch ev.Kind {
	case game.EventCast, game.EventSpellCopied:
		return ev.CardID
	default:
		return ev.StackItemID
	}
}

// putOntoTheStackWithASingleTarget is Grip of Chaos' condition: the
// object just put onto the stack has exactly one target.
func putOntoTheStackWithASingleTarget(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	id := putOntoTheStackItem(ev)
	return id != uuid.Nil && g.StackItemTargetCountForEffect(id) == 1
}

// gripReselectsTheTarget is Grip of Chaos' resolution: the intervening
// if again, then the random reselection. The object that was put onto
// the stack is named by the trigger's own triggering event
// (StackItem.Trigger); one that is no longer on the stack is not
// reselected, and neither is one that no longer has a single target.
func gripReselectsTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := putOntoTheStackItem(ctx.Trigger().Event)
	if id == uuid.Nil || g.StackItemTargetCountForEffect(id) != 1 {
		return nil
	}
	err := g.ReselectTargetsAtRandomForEffect(game.RandomRetarget{
		ItemID: id,
		Drawer: item.Controller,
		Source: item.SourceCardID,
	})
	if err == game.ErrCardNotFound {
		return nil
	}
	return err
}
