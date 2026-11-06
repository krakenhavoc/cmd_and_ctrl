package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ceddea59-de47-4ddd-b424-6e1dcb13ab79",
		Name:         "Allied Teamwork",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Allied Teamwork — create a 1/1 white Ally creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Ally"}, YoursOnly: true}, 1, 1),
		},
	})
}
