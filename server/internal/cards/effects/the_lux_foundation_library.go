package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Lux Foundation Library — Plane — The Library (Doctor Who, #4870):
//
//	"Players have no maximum hand size.
//	 Whenever a creature you control deals combat damage to a player, you
//	 may draw a card.
//	 Whenever chaos ensues, put a shadow counter on target creature. (A
//	 creature with shadow can block or be blocked by only creatures with
//	 shadow.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The combat damage ability triggers once per creature and damage dealt
// (each creature's damage is its own event, CR 603.2). The chaos ability
// (CR 311.7) puts a shadow counter, a keyword counter (CR 122.1b), on any
// creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68750b81-1309-4362-84b3-8835b3d71d28",
		Name:         "The Lux Foundation Library",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{PlayersHaveNoMaxHandSize()},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "The Lux Foundation Library — draw a card", Do(DrawCards{N: 1})), "The Lux Foundation Library — draw a card?", 1),
			Targeting(WheneverChaosEnsues("The Lux Foundation Library — put a shadow counter on target creature",
				putACounterOnTheTarget(game.CounterShadow)),
				TargetCreature("target creature")),
		},
	})
}
