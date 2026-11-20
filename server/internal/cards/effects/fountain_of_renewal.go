package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fountain of Renewal — Artifact {1} (Core Set 2019, #5527):
//
//	"At the beginning of your upkeep, you gain 1 life.
//	{3}, Sacrifice this artifact: Draw a card."
//
// Nyx-Fleece Ram's upkeep life, and a sacrifice ability that draws a
// card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df024c55-c008-48ec-a1a5-02ce336e3de6",
		Name:         "Fountain of Renewal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Fountain of Renewal — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{3}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
