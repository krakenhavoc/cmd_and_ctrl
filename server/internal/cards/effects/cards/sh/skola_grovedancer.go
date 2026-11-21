package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skola Grovedancer:
//
//	"Whenever a land card is put into your graveyard from anywhere, you
//	 gain 1 life.
//	 {2}{G}: Mill a card."
//
// One trigger per land card (not a batch): each land card put into the
// graveyard gains 1 life, even in the same milling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e3246ef-e0b4-49bc-a715-b69c4b6be461",
		Name:         "Skola Grovedancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventMill, game.EventZoneMove, game.EventLTB}, effects.Cmdr2B13LandCardToYourGraveyard,
				"Skola Grovedancer — gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Mill a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect:  effects.Do(effects.MillCards{N: 1}),
		}},
	})
}
