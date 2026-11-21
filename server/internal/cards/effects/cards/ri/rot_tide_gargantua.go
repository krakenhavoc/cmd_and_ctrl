package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "69c54156-255b-40d9-9c63-4cb9099aa859",
		Name:         "Rot-Tide Gargantua",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.Exploit("Rot-Tide Gargantua", rotTideGargantuaEdictBody)),
		},
	})
}

// rotTideGargantuaEdictBody is "When this creature exploits a creature,
// each opponent sacrifices a creature of their choice." An on-disk key:
// never renamed, never reused.
var rotTideGargantuaEdictBody = game.ReflexiveBody("rot-tide-gargantua/edict", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"}.Apply(effects.NewContext(g, item))
}), nil)
