package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mutant Surveyor — Creature — Mutant Scout {2}{B}, 2/3:
//
//	"Start your engines!
//	 {2}: This creature gets +1/+1 until end of turn.
//	 Max speed — {3}, Exile this card from your graveyard: Draw a card."
//
// ADR 0138 (#2711). The pump reads its source once at resolution and
// does nothing if it left the battlefield; the graveyard ability is the
// shared Surveyor body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f35aace7-57b5-44b8-91d2-4710eb7d5548",
		Name:            "Mutant Surveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: This creature gets +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
				Cost:    effects.ManaCost("{2}"),
				Effect:  effects.ThisCreatureUntilEOT("Mutant Surveyor — +1/+1 until end of turn", 1, 1),
			},
			effects.MaxSpeedSurveyorDraw(),
		},
	})
}
