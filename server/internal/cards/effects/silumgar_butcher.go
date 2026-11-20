package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silumgar Butcher — Creature — Zombie Djinn {4}{B}, 3/3 (Dragons of Tarkir,
// #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, target creature gets -3/-3 until end of turn."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff
// is a reflexive trigger raised only when a creature was really
// sacrificed, and a targeted payoff chooses its target then.
// The payoff is Aphetto Exterminator's -3/-3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77216efa-4f34-4534-9df3-918f0b888ce1",
		Name:         "Silumgar Butcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Exploit("Silumgar Butcher", silumgarButcherShrinkBody),
		},
	})
}

// silumgarButcherShrinkBody is "When this creature exploits a creature, target creature gets -3/-3 until end of turn." An on-disk key:
// never renamed, never reused.
var silumgarButcherShrinkBody = game.ReflexiveBody("silumgar-butcher/minus-three", simpleBody(boostTargetUntilEOT("Silumgar Butcher — -3/-3 until end of turn", -3, -3)),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
