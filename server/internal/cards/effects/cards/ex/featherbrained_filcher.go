package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Featherbrained Filcher — {W} Creature — Bird Mutant 0/2 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Flying
//	 When this creature leaves the battlefield, create a Food token."
//
// City Pigeon's leaves trigger (a look-back trigger, CR 603.10a) and the
// CR 111.10b Food token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79a9fc1c-a6a9-483a-9ba7-d09fe41760c3",
		Name:            "Featherbrained Filcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Featherbrained Filcher — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
