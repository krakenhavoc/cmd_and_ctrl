package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attractions.go is the card vocabulary of Attractions (CR 717, 701.51,
// 701.52, 702.159; ADR 0207). The engine runs the keyword actions
// (game.OpenAttractionsForEffect, game.RollToVisitForEffect), the
// turn-based roll as each precombat main phase begins, the Attraction
// deck and the junkyard; a card only has to say what it prints:
//
//	OpenAttractions{N: 1}            "open an Attraction" / "open two Attractions"
//	OpenAnAttraction                 the same, as a whole Effect
//	WhenThisEntersOpenAnAttraction   "When this creature enters, open an Attraction."
//	RollToVisitYourAttractions{}     "roll to visit your Attractions"
//	Visit(name, effect)              an Attraction's "Visit — <effect>" (CR 702.159a)
//	Attraction()                     the predicate "Attraction" ("destroy target Attraction")
//
// A step that opens an Attraction whose entry asks a question (a CR 614
// window) holds the steps after it, like every other step (do_walk.go).

// OpenAttractions is "<player> opens N Attractions" (CR 701.51): N
// times, the top card of their Attraction deck onto the battlefield
// under their control. Player defaults to the effect's controller; N
// below one is one. A player with no Attraction deck, or an empty one,
// opens nothing (CR 701.51a); that is not an error.
type OpenAttractions struct {
	Player uuid.UUID
	N      int
}

func (o OpenAttractions) Apply(ctx *Context) error {
	player := o.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.OpenAttractionsForEffect(player, ctx.Source(), max(o.N, 1))
}

// OpenAnAttraction is "open an Attraction" as a whole Effect, for a
// trigger whose text is only that.
var OpenAnAttraction = Do(OpenAttractions{N: 1})

// openLabel is the stack label of a card's open ability.
func openLabel(name string) string { return name + " — open an Attraction" }

// WhenThisEntersOpenAnAttraction is "When this creature enters, open an
// Attraction." — the Employees' common line.
func WhenThisEntersOpenAnAttraction(name string) game.TriggeredAbility {
	return WhenThisEnters(openLabel(name), OpenAnAttraction)
}

// RollToVisitYourAttractions is "roll to visit your Attractions" (CR
// 701.52a): a d6, then every Attraction the player controls with that
// number lit up is visited. Player defaults to the effect's controller.
// A player who controls no Attraction still rolls (the die is still a
// die roll, so "whenever you roll a 1" sees it) and visits nothing.
type RollToVisitYourAttractions struct {
	Player uuid.UUID
}

func (r RollToVisitYourAttractions) Apply(ctx *Context) error {
	player := r.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	if ctx.Game.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return ctx.Game.RollToVisitForEffect(player, ctx.Source())
}

// thisAttractionVisited is the trigger condition of a visit ability (CR
// 702.159a): "whenever you roll to visit your Attractions, if the result
// is equal to a number that is lit up on this Attraction". The engine
// announces EventVisitAttraction only for an Attraction whose lights
// match the roll, so the event naming this object, rolled by its
// controller, is the whole condition. The intervening "if" cannot turn
// false before the ability resolves: the lights are printed on the card
// and nothing changes them.
func thisAttractionVisited(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.Kind == game.EventVisitAttraction &&
		ev.CardID == source.InstanceID && ev.Actor == source.Controller
}

// Visit is an Attraction's visit ability, "Visit — <effect>" (CR
// 702.159a). The label is the stack label.
func Visit(name string, effect Effect) game.TriggeredAbility {
	return On(game.EventVisitAttraction, thisAttractionVisited, name+" — visit", effect)
}

// Attraction is the predicate "Attraction": a permanent with the
// Attraction subtype now (CR 717.1, 205.3g).
func Attraction() CardPredicate { return Subtype(game.AttractionSubtype) }
