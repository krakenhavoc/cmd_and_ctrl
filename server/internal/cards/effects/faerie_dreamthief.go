package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Dreamthief — Creature — Faerie Warlock {B}, 1/1 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 When this creature enters, surveil 1. (Look at the top card of your
//	 library. You may put it into your graveyard.)
//	 {2}{B}, Exile this card from your graveyard: You draw a card and you
//	 lose 1 life."
//
// The entry trigger is Clandestine Meddler's surveil 1. The graveyard
// ability is Bonebind Orator's exile-this cost (ExileThis, working only
// from the graveyard, at instant speed) with Greta, Sweettooth Scourge's
// draw-and-lose-1 body. It targets nothing, so it declares what it
// answers (value).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6e68414-80cb-4e08-ae84-b084a3f4b6de",
		Name:            "Faerie Dreamthief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Faerie Dreamthief — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Exile this card from your graveyard: You draw a card and you lose 1 life.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  Do(DrawCards{N: 1}, GainLife{Amount: -1}),
		}},
	})
}
