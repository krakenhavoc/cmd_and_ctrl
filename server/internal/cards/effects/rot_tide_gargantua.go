package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rot-Tide Gargantua — Creature — Zombie Kraken {3}{B}{B}, 5/4:
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, each opponent sacrifices a
//	 creature of their choice."
//
// Exploit is the shared constructor (exploit.go); the payoff is an
// untargeted reflexive trigger, and the edict is Kaya's Guile's
// (EachPlayerSacrifices, each opponent choosing among their own
// creatures as it resolves). An opponent with no creature sacrifices
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69c54156-255b-40d9-9c63-4cb9099aa859",
		Name:         "Rot-Tide Gargantua",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(Exploit("Rot-Tide Gargantua", rotTideGargantuaEdictBody)),
		},
	})
}

// rotTideGargantuaEdictBody is "When this creature exploits a creature,
// each opponent sacrifices a creature of their choice." An on-disk key:
// never renamed, never reused.
var rotTideGargantuaEdictBody = game.ReflexiveBody("rot-tide-gargantua/edict", simpleBody(func(g *game.Game, item *game.StackItem) error {
	return EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"}.Apply(NewContext(g, item))
}), nil)
