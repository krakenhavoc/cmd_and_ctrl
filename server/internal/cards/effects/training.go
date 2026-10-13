package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// training.go — training (CR 702.138): "Whenever this creature attacks
// with another creature with greater power, put a +1/+1 counter on this
// creature." Hopeful Initiate (#3172).
//
// The condition is read as the attack is declared, from the power of the
// creatures that are attacking at that moment, and is not an
// intervening "if": the printed reminder says "whenever ... attacks
// with", so once the trigger is on the stack the counter goes on
// whatever happened to the bigger creature in response. The counter
// itself is skipped if the trainee has left or come back as a new
// object (CR 400.7).

// Training is "Training" on a creature: the shared trigger, built once
// so every training card compares power the same way.
//
//	Triggered: []game.TriggeredAbility{Training("Hopeful Initiate — training")},
func Training(label string) game.TriggeredAbility {
	return On(game.EventAttack, AllOf(ThisAttacked, attacksWithGreaterPower), label, trainingCounter)
}

// attacksWithGreaterPower is training's condition: another creature the
// source's controller has declared as an attacker has greater power than
// the source (CR 702.138a).
func attacksWithGreaterPower(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if g == nil {
		return false
	}
	mine := source.CurrentPower()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == source.InstanceID || c.Controller != source.Controller || c.AttackingTarget == uuid.Nil {
			continue
		}
		if c.CurrentPower() > mine {
			return true
		}
	}
	return false
}

// trainingCounter puts the +1/+1 counter on the trainee.
func trainingCounter(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	return g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, 1)
}
