package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quakebringer — Creature — Giant Berserker {3}{R}{R}, 5/4:
//
//	"Your opponents can't gain life.
//	 At the beginning of your upkeep, Quakebringer deals 2 damage to
//	 each opponent. This ability triggers only if Quakebringer is on
//	 the battlefield or if Quakebringer is in your graveyard and you
//	 control a Giant.
//	 Foretell {2}{R}{R}"
//
// "Your opponents can't gain life" is ADR 0107 §5's battlefield static
// (CR 119.7, #1880); it applies only while Quakebringer is on the
// battlefield, as every static does.
//
// The upkeep ability functions from two zones (CR 113.6). The engine
// declares a trigger's zones as a list in which the battlefield is the
// empty default, so the one printed ability is written as two rows
// with the same effect: the ordinary battlefield one, and a
// graveyard one (TriggeredAbility.Zones) whose condition is "you
// control a Giant" — checked when it would trigger, which is what
// "triggers only if" says (not an intervening if). Only one of the two
// can ever fire, because the card is in one zone at a time. From the
// graveyard the card itself is the damage source, as printed.
//
// Foretell is the shared special action.
//
// No simplification.
func init() {
	const label = "Quakebringer — 2 damage to each opponent"
	fromGraveyard := On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Owner && quakebringerControlsAGiant(g, source.Owner)
	}, label+" (from your graveyard)", quakebringerDamage)
	fromGraveyard.Zones = []game.ZoneKind{game.ZoneGraveyard}
	Register(Spec{
		OracleID:       "bbe41a21-f8ad-4e5c-88d8-2a41dfd32d13",
		Name:           "Quakebringer",
		Completeness:   CompletenessFull,
		CantGainLife:   OpponentsCantGainLife(),
		SpecialActions: []game.SpecialAction{Foretell("{2}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep(label, quakebringerDamage),
			fromGraveyard,
		},
	})
}

// quakebringerDamage is the upkeep ability's body: 2 damage to each
// opponent, Quakebringer the source wherever it is.
func quakebringerDamage(g *game.Game, item *game.StackItem) error {
	return damageToEachOpponent(g, item, 2)
}

// quakebringerControlsAGiant reports whether `player` controls a Giant
// on the battlefield.
func quakebringerControlsAGiant(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype("Giant") {
			return true
		}
	}
	return false
}
