package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// first_life_loss.go — "whenever you lose life for the first time each
// turn" (#2540, Gonti's Machinations).
//
// The turn tally's LifeLost sums every loss of the turn (a negative
// EventChangeLife, and the life an EventDealDamage cost, #2105). The
// tally listener is registered ahead of the trigger harvester, so when
// a trigger's AppliesTo runs for a loss, LifeLost ALREADY includes that
// loss: TestATriggerSeesTheTallyIncludingTheLossItIsAskedAbout pins it.
// A loss is therefore the turn's first exactly when the tally equals
// it, because every earlier loss would have added to the tally.
//
// That also makes a burst of simultaneous losses one trigger. Several
// creatures dealing combat damage to one player are one event each, the
// first sees only itself, and every later one sees the running total
// and is not first. No batch bookkeeping is needed.
//
// The tally belongs to the PLAYER, not to the ability's source, so a
// Gonti's Machinations that enters after the turn's first loss never
// sees a "first" one (CR 603.2: it can only trigger on events it was on
// the battlefield for).

// youLostLifeForTheFirstTimeThisTurn reports whether ev is `controller`
// losing life, and how much, and whether it is the first loss of the
// turn. A gain is not a loss; damage that cost no life (infect,
// prevention, a locked life total) is not a loss either, so it neither
// triggers nor uses up the turn's first.
func youLostLifeForTheFirstTimeThisTurn(ev game.Event, controller uuid.UUID, g *game.Game) (lost int, first bool) {
	lost, ok := s22PlayerLostLife(ev, controller, g)
	if !ok {
		return 0, false
	}
	return lost, g.TurnTallyFor(controller).LifeLost == lost
}

// WheneverYouLoseLifeForTheFirstTimeEachTurn is the trigger itself. The
// effect runs once per turn at most, whatever the loss was made of.
func WheneverYouLoseLifeForTheFirstTimeEachTurn(label string, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			_, first := youLostLifeForTheFirstTimeThisTurn(ev, source.Controller, g)
			return first
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label)
		},
		Effect: effect,
	}
}
