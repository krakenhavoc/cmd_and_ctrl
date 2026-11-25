package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serpent of Yawning Depths — Enchantment Creature — Serpent {4}{U}{U}, 6/6
// (Secret Lair Drop, #5462):
//
//	"Krakens, Leviathans, Octopuses, and Serpents you control can't be
//	 blocked except by Krakens, Leviathans, Octopuses, and Serpents."
//
// One block rule (CR 509.1b, Midnight Pathlighter's shape): an attacker that
// is a Kraken, Leviathan, Octopus or Serpent its controller controls can be
// blocked only by a creature that is one of those four types. The creature
// types are read through the layers, so a changeling counts on both sides, and
// the Serpent itself is covered.
//
// No simplification.
func init() {
	seaMonster := effects.Or(effects.OfCreatureType("Kraken"), effects.OfCreatureType("Leviathan"), effects.OfCreatureType("Octopus"), effects.OfCreatureType("Serpent"))
	effects.Register(effects.Spec{
		OracleID:     "d31c6043-e005-4efe-9aa7-d82cb122eebe",
		Name:         "Serpent of Yawning Depths",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnMatching(effects.And(effects.Creature(), effects.YouControl(), seaMonster)), seaMonster, "Krakens, Leviathans, Octopuses, and Serpents"),
		},
	})
}
