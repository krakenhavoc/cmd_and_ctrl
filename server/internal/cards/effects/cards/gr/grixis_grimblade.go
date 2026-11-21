package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grixis Grimblade — Creature — Zombie Warrior {U/R}{B}, 2/1 (Alara
// Reborn, #5586):
//
//	"As long as you control another multicolored permanent, this creature
//	 gets +1/+1 and has deathtouch. (Any amount of damage it deals to a
//	 creature is enough to destroy that creature.)"
//
// The blade cycle's shared static (arbB1Blade): a live +1/+1 and a
// deathtouch grant on the source alone, read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04b1280a-b084-4f65-a3f4-2ba41d6c26fa",
		Name:         "Grixis Grimblade",
		Completeness: effects.CompletenessFull,
		Static:       effects.ArbB1Blade("deathtouch"),
	})
}
