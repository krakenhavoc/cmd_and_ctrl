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
// Mandatory; Build deals 2 damage to the controller, sourced by the
// Vortex itself so combat-style damage routing is consistent.
func init() {
	Register(Spec{
		OracleID: "7652f328-e142-494b-a869-772ced10c26a",
		Name:     "Sulfuric Vortex",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				ctx := NewContext(g, nil)
				_ = DealDamage{
					Source: source.InstanceID,
					Target: source.Controller,
					Amount: 2,
				}.Apply(ctx)
				return nil
			},
		}},
	})
}
