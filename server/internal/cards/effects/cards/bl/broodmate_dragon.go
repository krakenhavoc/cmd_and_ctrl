package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broodmate Dragon — Creature — Dragon {3}{B}{R}{G}, 4/4 (Shards of
// Alara, #5588):
//
//	"Flying
//	 When this creature enters, create a 4/4 red Dragon creature token
//	 with flying."
//
// Dragon Trainer's enters trigger on a flier (CR 702.9b): the token is
// the shared 4/4 red flying Dragon row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86529f12-6f0f-4b0d-96f1-785474f6763a",
		Name:            "Broodmate Dragon",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Broodmate Dragon — create a 4/4 red Dragon with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 1})),
		},
	})
}
