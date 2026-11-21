package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brood Weaver — Creature — Spider {3}{G}, 2/4 (Innistrad: Midnight Hunt, S82):
//
//	"Reach
//	 When this creature dies, create a 1/2 green Spider creature token with reach."
//
// A dies trigger making the 1/2 Spider with reach.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f0a7ef59-7897-4e84-a85e-1fcbe3dc9f6a",
		Name:            "Brood Weaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Brood Weaver — create a 1/2 green Spider creature token with reach",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/2 green Spider with reach"), N: 1})),
		},
	})
}
