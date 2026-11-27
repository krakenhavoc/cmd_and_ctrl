package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foul Spirit — Creature — Spirit {2}{B}, 3/2 (Portal Second Age, S82 older sets batch 734, #5625):
//
//	"Flying
//	 When this creature enters, sacrifice a land."
//
// The land is chosen by its controller as the trigger resolves (CR 701.21); with no land there is nothing to sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8177d08-a603-42e9-9350-890eac05e184",
		Name:            "Foul Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Foul Spirit — sacrifice a land", effects.S82OldB614SacrificeALand("Foul Spirit — sacrifice a land")),
		},
	})
}
