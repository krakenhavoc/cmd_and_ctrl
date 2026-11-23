package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Havoc Jester — Creature — Devil {4}{R}, 5/5 (Ravnica: Clue Edition,
// #5929):
//
//	"Whenever you sacrifice a permanent, this creature deals 1 damage to
//	 any target."
//
// Mayhem Devil's trigger limited to its controller's sacrifices
// (cmdrB9's "Actor is the controller" read). A sacrifice is not a death:
// a destroyed permanent does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58f5bbd2-b876-4eb9-aa79-f4d3511961e0",
		Name:         "Havoc Jester",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventSacrifice},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventSacrifice && source != nil && ev.Actor == source.Controller
			},
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Key:     "Havoc Jester — deal 1 damage to any target",
			Effect:  effects.DamageToFirstTarget(1),
		}},
	})
}
