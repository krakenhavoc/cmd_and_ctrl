package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bridled Bighorn — Creature — Sheep Mount {3}{W}:
//
//	"Vigilance
//	 Whenever this creature attacks while saddled, create a 1/1 white
//	 Sheep creature token.
//	 Saddle 2"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a1ad848-6fda-46dd-aad8-2f6a5c71a696",
		Name:            "Bridled Bighorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Saddle(2)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Bridled Bighorn — create a 1/1 white Sheep", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Sheep"), N: 1})),
		},
	})
}
