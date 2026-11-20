package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Discordant Piper — Creature — Zombie Satyr {1}{B}, 2/1 (Theros Beyond Death, #5514):
//
//	"When this creature dies, create a 0/1 white Goat creature token."
//
// Woe Strider's Goat, made when the Piper dies rather than when it
// enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb0107f4-723a-4f2c-8e4b-237d5f668e3c",
		Name:         "Discordant Piper",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Discordant Piper — create a 0/1 white Goat",
				Do(CreateToken{Template: TokenCard("0/1 white Goat"), N: 1})),
		},
	})
}
