package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Allied Teamwork — Enchantment {2}{W} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"When this enchantment enters, create a 1/1 white Ally creature token.
//	 Allies you control get +1/+1."
//
// The token is the set's plain Ally row. The anthem is White Lotus
// Reinforcements' lord without "other", read through the layered view so
// a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ceddea59-de47-4ddd-b424-6e1dcb13ab79",
		Name:         "Allied Teamwork",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Allied Teamwork — create a 1/1 white Ally creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Ally"), N: 1})),
		},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Ally"}, YoursOnly: true}, 1, 1),
		},
	})
}
