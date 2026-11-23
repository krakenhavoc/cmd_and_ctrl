package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Generous Ent — Creature — Treefolk {5}{G}, 5/7 (The Lord of the Rings:
// Tales of Middle-earth):
//
//	"Reach
//	 When this creature enters, create a Food token.
//	 Forestcycling {1}"
//
// Balamb T-Rexaur's typecycling (Typecycling, CR 702.29: discard it to search
// for a Forest card) at {1}, and an enters trigger for the Food.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b6859c7a-4c4d-4652-938f-ecedce29d506",
		Name:            "Generous Ent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Generous Ent — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{1}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
