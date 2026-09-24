package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Indomitable — Legendary Artifact — Vehicle, {2}{U}{U}:
//
//	"Trample
//	 Whenever a creature you control deals combat damage to a player,
//	 draw a card.
//	 Crew 3
//	 You may cast this card from your graveyard as long as you control
//	 three or more tapped Pirates and/or Vehicles."
//
// The draw trigger is "a creature you control", not "this creature" or
// "one or more creatures" — one trigger per attacker that connects, no
// OncePerBatch, exactly the shape combatDamageToPlayerBy exists for
// (Bident of Thassa, Old Gnawbone).
//
// The graveyard cast is the CR 205.4e-family CastCondition/
// CastConditionLabel pair (ADR 0073 §7, Rakdos, Lord of Riots) rather
// than a caveat: the printed clause names a real, checkable board
// state, and the condition closure is what scopes it to the graveyard
// path ONLY, per the brief — a normal hand-cast (once the card is
// drawn or discarded there some other way) is never gated by it.
// CastCondition itself only sees (game, controller, card), not which
// zone the announce is coming from, so
// theIndomitableGraveyardCastCondition looks the card's own current
// zone up and returns true outright for anything other than the
// graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "276dc5c8-c8cf-4b9c-ad75-31876e6e040a",
		Name:               "The Indomitable",
		Completeness:       CompletenessFull,
		PrintedKeywords:    []string{"trample"},
		CastableZones:      []game.ZoneKind{game.ZoneGraveyard},
		CastCondition:      theIndomitableGraveyardCastCondition,
		CastConditionLabel: "You can't cast this from your graveyard unless you control three or more tapped Pirates and/or Vehicles.",
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "The Indomitable — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:  "Crew 3",
			Cost:   CrewCost(3),
			Effect: CrewEffect("The Indomitable"),
		}},
	})
}

// theIndomitableGraveyardCastCondition is "you control three or more
// tapped Pirates and/or Vehicles", scoped to the graveyard: a cast
// from any other zone (hand) is always permitted, because the printed
// clause only restricts the graveyard path.
func theIndomitableGraveyardCastCondition(g *game.Game, controller uuid.UUID, card game.Card) bool {
	z := g.FindCardZoneForEffect(card.InstanceID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return true
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.Tapped {
			continue
		}
		if isPirate(c) || c.HasSubtype("Vehicle") {
			n++
		}
	}
	return n >= 3
}
