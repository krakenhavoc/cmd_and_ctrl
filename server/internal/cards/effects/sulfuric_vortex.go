package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sulfuric Vortex — Enchantment for {1}{R}:
//
//	"At the beginning of your upkeep, Sulfuric Vortex deals 2 damage
//	to you.
//	If a player would gain life, that player gains no life instead."
//
// S19 sub-PR 5 implements the upkeep-damage trigger. The static
// "players can't gain life" clause is a continuous effect deferred
// to a later batch (it needs the life-gain replacement hook).
// Mandatory; on resolution the Vortex deals 2 damage to its
// controller, sourced by the Vortex itself so damage routing (and
// any future prevention) is consistent with combat-style damage.
func init() {
	Register(Spec{
		OracleID:     "7652f328-e142-494b-a869-772ced10c26a",
		Name:         "Sulfuric Vortex",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"players gain no life\" half isn't implemented; only the 2 damage to you each upkeep happens."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Sulfuric Vortex — 2 damage to you",
					func(g *game.Game, item *game.StackItem) error {
						return DealDamage{
							Source: item.SourceCardID,
							Target: item.Controller,
							Amount: 2,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
