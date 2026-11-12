package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3047ad53-1fbd-4b01-bf75-a3e633c863fe",
		Name:         "Diver Skaab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Exploit("Diver Skaab", diverSkaabTuckBody),
		},
	})
}

// diverSkaabTuckBody is "When this creature exploits a creature, target
// creature's owner puts it on their choice of the top or bottom of their
// library." An on-disk key: never renamed, never reused.
var diverSkaabTuckBody = game.ReflexiveBody("diver-skaab/tuck", simpleBody(func(g *game.Game, item *game.StackItem) error {
	return ownerTucksFirstTargetTopOrBottom(NewContext(g, item),
		"Diver Skaab — put it on your choice of the top or bottom of your library")
}), constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
