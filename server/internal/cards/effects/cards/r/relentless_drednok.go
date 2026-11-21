package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relentless Drednok — Artifact Creature — Robot {B}, 1/1 (Star Trek,
// #4973):
//
//	"Lifelink
//	 {2}{B}, Exile this card from your graveyard: You draw a card and
//	 lose 1 life."
//
// The second ability functions only from the graveyard (CR 113.6m), and
// exiling the card is part of its cost, so it can be activated once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "98cbac1d-7896-4c63-a265-03bc5d670fa6",
		Name:            "Relentless Drednok",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Exile this card from your graveyard: You draw a card and lose 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.B36DrawAndLoseOne,
		}},
	})
}
