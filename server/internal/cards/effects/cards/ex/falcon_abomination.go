package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "21eea63f-72b2-4155-bfad-f9937a8f8614",
		Name:            "Falcon Abomination",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Falcon Abomination — create a 2/2 Zombie with decayed", effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 1})),
		},
	})
}
