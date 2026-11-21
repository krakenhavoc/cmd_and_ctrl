package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diver Skaab — Creature — Zombie {3}{U}{U}, 3/5:
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, target creature's owner
//	 puts it on their choice of the top or bottom of their library."
//
// Exploit is the shared constructor (exploit.go): the payoff is a
// reflexive trigger raised only when a creature really went, and its
// target is chosen then, with the sacrifice already made. The tuck is
// Lagoon Breach's (ownerTucksFirstTargetTopOrBottom): the creature's
// OWNER chooses the end, and a target that is gone or illegal by
// resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3047ad53-1fbd-4b01-bf75-a3e633c863fe",
		Name:         "Diver Skaab",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Diver Skaab", diverSkaabTuckBody),
		},
	})
}

// diverSkaabTuckBody is "When this creature exploits a creature, target
// creature's owner puts it on their choice of the top or bottom of their
// library." An on-disk key: never renamed, never reused.
var diverSkaabTuckBody = game.ReflexiveBody("diver-skaab/tuck", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.OwnerTucksFirstTargetTopOrBottom(effects.NewContext(g, item),
		"Diver Skaab — put it on your choice of the top or bottom of your library")
}), effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))
