package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// becomes_crewed.go — "whenever this Vehicle becomes crewed" (CR
// 702.122e, #5671): "Whenever a crew ability of [this Vehicle] resolves."
//
// The engine says it with game.EventBecameCrewed, emitted after a crew
// ability's effect while the Vehicle is still on the battlefield as the
// object the ability came from, so a trigger built here goes on the stack
// once the Vehicle is already an artifact creature. A crew ability that
// was countered made nothing become crewed. It is NOT
// WhenThisCrewsOrSaddles, which is the creature's side of the payment and
// fires per creature while the cost is paid.
//
//	Printed                                             Written
//	Whenever this Vehicle becomes crewed, …             WhenBecomesCrewed(label, effect)
//	… becomes crewed for the first time each turn, …    WhenBecomesCrewedFirstTimeEachTurn(label, effect)
//	a creature that crewed it this turn                 CrewedThis(spec) / CrewersOfThis(ctx)

// WhenBecomesCrewed is "Whenever this Vehicle becomes crewed, …": once
// for each crew ability of the source that resolves, however many
// creatures paid for it. Set Targets on the result (or wrap it in
// Targeting) for a targeted trigger.
func WhenBecomesCrewed(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventBecameCrewed},
		Key:       label,
		AppliesTo: thisBecameCrewed,
		Effect:    effect,
	}
}

// WhenBecomesCrewedFirstTimeEachTurn is "Whenever this Vehicle becomes
// crewed for the first time each turn, …" (Mindlink Mech). The event
// carries how many crew abilities of this object have resolved this
// turn, so the condition is about the Vehicle becoming crewed and not
// about this ability having triggered: a first crew whose trigger found
// no legal target still used the turn's one chance.
func WhenBecomesCrewedFirstTimeEachTurn(label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	t := WhenBecomesCrewed(label, effect)
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return thisBecameCrewed(ev, source, lki, g) && ev.Amount == 1
	}
	return t
}

// thisBecameCrewed matches the EventBecameCrewed of the source itself.
func thisBecameCrewed(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.CardID == source.InstanceID
}
