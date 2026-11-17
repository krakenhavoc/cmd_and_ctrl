package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// clockwork.go — the Clockwork creatures' shared clause (Mirrodin,
// #5606): "Whenever this creature attacks or blocks, remove a +1/+1
// counter from it at end of combat." Clockwork Beetle, Condor, Vorrac
// and Dragon print it word for word. Append-only.
//
// It is Dorothea, Vengeful Victim's trigger with a counter in place of
// the sacrifice: one trigger with two conditions, which fires once
// however many attackers the creature blocks (CR 509.3a), and whose
// resolution sets up a delayed trigger at the beginning of this
// combat's end of combat step (CR 511.2) whose source is this object.
// If the creature has left the battlefield by then, even to come back
// as a new object, the delayed trigger does nothing (CR 603.7c,
// CR 400.7); with no +1/+1 counter left there is nothing to remove.

// clockworkRemoveCounterBody is the delayed trigger "remove a +1/+1
// counter from it at end of combat" schedules. A registered key (ADR
// 0041 phase 3), so a table with it waiting is a restore point.
//
// Assigned in init, as dorotheaSacrificeBody is: a var initialiser
// would be an initialisation cycle through the counter primitives.
var clockworkRemoveCounterBody game.BodyRef

func init() {
	clockworkRemoveCounterBody = game.SimpleDelayedBody("clockwork/remove-a-plus-one-counter-at-end-of-combat",
		removePlusOneCounterFromDelayedTriggerSource)
}

// removePlusOneCounterFromDelayedTriggerSource removes one +1/+1
// counter from the delayed trigger's source (CR 603.7e), if that is
// still the same object on the battlefield.
func removePlusOneCounterFromDelayedTriggerSource(g *game.Game, item *game.StackItem) error {
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	return removeACounterFromThis(game.CounterPlusOne)(g, item)
}

// thisAttacksOrBlocks is "whenever this creature attacks or blocks":
// its own attack declaration, or the first of its blocks (CR 509.3a).
func thisAttacksOrBlocks(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
}

// clockworkLosesACounterAtEndOfCombat is the printed trigger, labelled
// with the card's name.
func clockworkLosesACounterAtEndOfCombat(name string) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, thisAttacksOrBlocks,
		name+" — remove a +1/+1 counter from it at end of combat",
		func(g *game.Game, item *game.StackItem) error {
			return ScheduleDelayedTrigger{
				At:    game.StepEndCombat,
				Label: name + " — remove a +1/+1 counter from it",
				Body:  clockworkRemoveCounterBody,
			}.Apply(NewContext(g, item))
		})
}
