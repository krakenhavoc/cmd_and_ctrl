package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Disturbance — Enchantment — Aura Curse {2}{B} (Commander 2017,
// #5533):
//
//	"Enchant player
//	 Whenever enchanted player is attacked, create a 2/2 black Zombie
//	 creature token. Each opponent attacking that player does the same."
//
// Curse of Opulence's shape with a Zombie in place of a Gold token
// (c17CurseWhenAttacked): one declaration is one "is attacked", and each
// attacking opponent creates one Zombie under their own control however
// many creatures they send.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cbd36d9-de47-41b0-9ef4-a72ca01adccd",
		Name:         "Curse of Disturbance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: effects.C17CurseWhenAttacked("Curse of Disturbance — create a 2/2 black Zombie",
			"Curse of Disturbance — the attacking opponent creates a 2/2 black Zombie",
			effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1})),
	})
}
