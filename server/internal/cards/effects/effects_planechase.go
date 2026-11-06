package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// effects_planechase.go is the card side of Planechase (ADR 0216 §§5–6,
// slice 2, #4870): the shapes a plane card's abilities take. A plane is
// an ordinary catalog card keyed by its oracle ID; its abilities function
// while it is face up in the command zone (CR 901.7), where the engine's
// planar walks find them (game/planechase_sources.go). Append-only: a
// shape two planes share goes here, never in a card file.
//
// The face-up card's controller is the planar controller, so "you" below
// is read off the source's Controller like any permanent's.

// WheneverYouRollChaos is a plane's chaos ability, "Whenever you roll
// {CHAOS}, <effect>" (CR 901.9b): the planar die showed the chaos symbol
// for the source's controller. The roll is the triggering event, so a
// table with the trigger waiting is a restore point.
func WheneverYouRollChaos(label string, effect func(*game.Game, *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventRollDie}, Key: label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.PlanarFace == game.PlanarFaceChaos && ev.Actor == source.Controller
		},
		Effect: effect,
	}
}

// WhenYouPlaneswalkToThis is "When you planeswalk to <this plane>,
// <effect>" (CR 901.11): the source was just turned face up by its
// controller's planeswalk. The starting plane does not trigger it (CR
// 901.5).
func WhenYouPlaneswalkToThis(label string, effect func(*game.Game, *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventPlaneswalkTo}, Key: label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Actor == source.Controller
		},
		Effect: effect,
	}
}

// WhenYouPlaneswalkAwayFromThis is "When you planeswalk away from <this
// plane>, <effect>" (CR 901.11b): the harvester looks back at the card
// that has just left the face-up set.
func WhenYouPlaneswalkAwayFromThis(label string, effect func(*game.Game, *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventPlaneswalkFrom}, Key: label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Actor == source.Controller
		},
		Effect: effect,
	}
}

// UntilYouPlaneswalkAway is the duration "until you planeswalk away from
// <this plane>" for the face-up planar card `plane` (CR 901.11), or ok
// false when it is no longer face up, in which case the effect is not
// started (CR 611.2b).
func UntilYouPlaneswalkAway(g *game.Game, plane uuid.UUID) (game.Duration, bool) {
	return g.UntilPlaneswalkAwayDuration(plane)
}
