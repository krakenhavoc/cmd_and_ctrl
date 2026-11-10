package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0e3246ef-e0b4-49bc-a715-b69c4b6be461",
		Name:         "Skola Grovedancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB}, cmdr2B13LandCardToYourGraveyard,
				"Skola Grovedancer — gain 1 life", Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: Mill a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{G}"),
			Effect:  Do(MillCards{N: 1}),
		}},
	})
}
