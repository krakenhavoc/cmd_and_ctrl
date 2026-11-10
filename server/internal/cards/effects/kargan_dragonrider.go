package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kargan Dragonrider — Creature — Human Warrior {1}{R}, 2/2 (Foundations,
// #4465):
//
//	"As long as you control a Dragon, this creature has flying."
//
// Croakid Amphibonaut's conditional grant (selfWhileYouControlA) with a
// creature-type test (Subtype "Dragon", read through the layers), read at
// every layer pass, so flying comes and goes with the Dragon.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df366871-8ed7-4183-8b0c-414ca361a711",
		Name:         "Kargan Dragonrider",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(selfWhileYouControlA(Subtype("Dragon")), "flying"),
		},
	})
}
