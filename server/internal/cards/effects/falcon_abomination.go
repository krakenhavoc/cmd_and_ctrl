package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falcon Abomination — Creature — Zombie Bird {2}{U}, 2/2 (EDHREC rank
// 18744):
//
//	"Flying
//	 When this creature enters, create a 2/2 black Zombie creature
//	 token with decayed. (It can't block. When it attacks, sacrifice
//	 it at end of combat.)"
//
// The first decayed card (#2650). Decayed rides the token template as
// a keyword; the engine derives its "can't block" and its attack
// trigger from the ability list (game/decayed.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "21eea63f-72b2-4155-bfad-f9937a8f8614",
		Name:            "Falcon Abomination",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Falcon Abomination — create a 2/2 Zombie with decayed", Do(CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 1})),
		},
	})
}
