package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heir of Falkenrath // Heir to the Night — Creature — Vampire {1}{B}, 2/1 //
// Creature — Vampire Berserker 3/2 (Shadows over Innistrad, S82 OLD):
//
//	Front: "Discard a card: Transform this creature. Activate only once each
//	        turn."
//	Back:  "Flying"
//
// The once-a-turn limit reads the activation tally (CR 602.1b), so an
// activation still on the stack counts. The discard is the cost.
//
// No simplification.
func init() {
	const oracle = "d5b9c62d-fe98-4cbc-b4b6-012702bf6657"
	const label = "Discard a card: Transform this creature. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Heir of Falkenrath",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.DiscardACard(),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Heir to the Night",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
