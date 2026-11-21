package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeering Homunculus — Creature — Homunculus, {1}{U}, 0/4:
//
//	"When this creature enters, you may goad target creature.
//	(Until your next turn, that creature attacks each combat if able
//	and attacks a player other than you if able.)"
//
// #1599: an "attack requirement" catalog card built entirely on
// GoadTarget (goad.go) — the ETB is optional (CR 603.5's "you may"),
// mandatory once accepted, with the ordinary structured target clause
// so an opponent can remove the chosen creature in response and the
// trigger fizzles (CR 608.2b) rather than goading nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "146de263-fece-4a00-ab67-35dc934165d8",
		Name:         "Jeering Homunculus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(game.TriggeredAbility{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets:   effects.TargetCreature("target creature"),
				Key:       "Jeering Homunculus — goad target creature",
				Effect:    effects.GoadTarget,
			}, "Jeering Homunculus — goad target creature?"),
		},
	})
}
