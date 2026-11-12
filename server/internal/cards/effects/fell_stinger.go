package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fell Stinger — Creature — Zombie Scorpion {2}{B}, 3/2:
//
//	"Deathtouch
//	 Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, target player draws two
//	 cards and loses 2 life."
//
// Deathtouch is a printed keyword. Exploit is the shared constructor
// (exploit.go); the payoff is Sign in Blood's body on a reflexive
// trigger, whose target player (any player, its controller included) is
// chosen once a creature has really been sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18dd694f-e239-4603-85fd-b34b028ff6a4",
		Name:            "Fell Stinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Exploit("Fell Stinger", fellStingerDrawBody),
		},
	})
}

// fellStingerDrawBody is "When this creature exploits a creature, target
// player draws two cards and loses 2 life." An on-disk key: never
// renamed, never reused.
var fellStingerDrawBody = game.ReflexiveBody("fell-stinger/draw-and-lose", simpleBody(func(g *game.Game, item *game.StackItem) error {
	return signInBloodOnResolve(item, NewContext(g, item))
}), constTargets(func() *game.TargetSpec { return TargetPlayer("target player") }))
