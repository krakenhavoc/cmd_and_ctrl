package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Attended Knight — Creature — Human Knight {2}{W}, 2/2 (Magic 2013, #5568):
//
//	"First strike
//	 When this creature enters, create a 1/1 white Soldier creature
//	 token."
//
// Argivian Cavalier's enters trigger on a first striker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3eb8f65-5eba-42a9-a1b4-4f37b14d03d9",
		Name:            "Attended Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Attended Knight — create a 1/1 white Soldier", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
