package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garrison Sergeant — Creature — Lizard Soldier {3}{R}{W}, 3/3 (Guilds of
// Ravnica, #5525):
//
//	"This creature has double strike as long as you control a Gate."
//
// Armory Guard's self-only keyword grant (layer 6, CR 613.1f) with double
// strike: the Gates are read live, through the layers, so a Gate that is turned
// into something else or leaves ends it.
//
// No simplification.
func init() {
	doubleStrike := effects.KeywordGrant(effects.SelfWhileYouControlA(effects.HasSubtype("Gate")), "double strike")
	doubleStrike.Label = "This creature has double strike as long as you control a Gate."
	effects.Register(effects.Spec{
		OracleID:     "cdf8b698-57d6-4d17-8168-87680060c0d6",
		Name:         "Garrison Sergeant",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{doubleStrike},
	})
}
