package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nutrient Block — Artifact {1} (Edge of Eternities):
//
//	"Indestructible (Effects that say "destroy" don't destroy this
//	 artifact.)
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// Chromatic Star's shape with a life-gain ability: the activation's three
// cost parts are paid together, and the draw is a leaves-for-the-
// graveyard trigger on the artifact itself, so it fires however the
// artifact got there (sacrificed to this ability, to another effect, or
// state-based). Indestructible is the printed keyword, read through the
// layers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04a0382f-7e4e-4450-ae70-18999379d875",
		Name:            "Nutrient Block",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect:  Do(GainLife{Amount: 3}),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Nutrient Block — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
