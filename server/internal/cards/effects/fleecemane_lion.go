package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fleecemane Lion — 3/3 Cat for {G}{W}:
//
//	"{3}{G}{W}: Monstrosity 1. (If this creature isn't monstrous, put
//	 a +1/+1 counter on it and it becomes monstrous.)
//	 As long as this creature is monstrous, it has hexproof and
//	 indestructible."
//
// The "as long as it's monstrous" static is the card (#1700): a layer-6
// self-grant behind the Monstrous gate, so before the designation the
// Lion is an ordinary 3/3 anyone can target and destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3c8cdf7-a26a-45eb-9498-99b618cfba99",
		Name:         "Fleecemane Lion",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Monstrosity(ManaCost("{3}{G}{W}"), 1)},
		Static:       []game.StaticAbility{MonstrousKeywords("hexproof", "indestructible")},
	})
}
