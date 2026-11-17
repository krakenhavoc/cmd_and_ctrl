package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_activation_count_rider.go — "If this ability has been activated
// four or more times this turn, sacrifice this creature at the beginning
// of the next end step" on a mana ability (#5745): Farrelite Priest's and
// Initiates of the Ebon Hand's "{1}: Add {W}" and "{1}: Add {B}".
//
// # What "this ability" counts
//
// The count is game.ActivatedThisTurn(source, label), the per-turn half
// of the activation record (game/activation_tally.go). It is keyed by
// the OBJECT and the printed ability, and both of those are the rule's:
//
//   - per OBJECT, because a permanent that leaves the battlefield and
//     comes back is a new object with no memory of its previous existence
//     (CR 400.7). A flickered Priest starts again from zero, and a token
//     copy of one has its own count.
//   - per ABILITY, because the sentence counts "this ability", not the
//     creature's other abilities.
//   - per TURN, emptied on the turn advance with the rest of the turn's
//     record. It is not per controller: an Initiates that changes control
//     keeps its count, since control is not a zone change.
//
// The record is written for every mana activation on both of the paths
// that make one: ActivateManaAbility (a click, a bot's move) and the
// auto-tapper's executor. Both write it before anything is paid, and the
// rider runs after the mana is added (CR 605.3b: the ability resolves
// immediately after it is activated), so the activation being resolved
// is already in the count: the fourth one reads 4.
//
// The auto-tapper never plans one of these abilities. This rider is a
// closure the planner cannot read (game.ManaAbilityShape.riderDeclared),
// and spending a creature the player did not choose to spend would be the
// wrong default anyway. The executor's write still covers it, so a future
// planner that learns the rider counts it the same way.
//
// # The sacrifice
//
// "Four or more", so every activation from the fourth on creates its own
// delayed triggered ability during its resolution (CR 603.7a). The first
// to resolve sacrifices the creature and the rest find nothing to act on.
// The delayed trigger's source is the creature as the object it is now
// (DelayedTrigger.SourceObject), so a creature flickered before the end
// step is a new object and is not sacrificed (CR 603.7c). Scheduled at
// game.StepEnd, so an activation during an end step waits for the next
// one: that step has already begun.

// sacrificeThisAfterActivationsBody is the delayed trigger's body: the
// trigger's source is sacrificed if it is still the same object on the
// battlefield under the trigger controller's control. A registered key
// (ADR 0041 phase 3), so a table with the trigger waiting is a restore
// point. Assigned in init, for dorotheaSacrificeBody's reason.
var sacrificeThisAfterActivationsBody game.BodyRef

func init() {
	sacrificeThisAfterActivationsBody = game.SimpleDelayedBody("activation-count/sacrifice-this-creature", sacrificeDelayedTriggerSource)
}

// SacrificeAtNextEndStepIfActivatedRider is the mana-ability rider "If
// this ability has been activated n or more times this turn, sacrifice
// this creature at the beginning of the next end step".
//
// abilityLabel must be the ability's own Label exactly: it is the key
// the activation record is written under, the contract every reader of
// game.ActivatedThisTurn keeps. triggerLabel is the delayed trigger's
// stack label.
func SacrificeAtNextEndStepIfActivatedRider(n int, abilityLabel, triggerLabel string) func(g *game.Game, controller, source uuid.UUID) error {
	return func(g *game.Game, controller, source uuid.UUID) error {
		if g.ActivatedThisTurn(source, abilityLabel) < n {
			return nil
		}
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   controller,
			SourceCardID: source,
			Label:        triggerLabel,
			At:           game.StepEnd,
			Body:         sacrificeThisAfterActivationsBody,
		})
		return nil
	}
}
