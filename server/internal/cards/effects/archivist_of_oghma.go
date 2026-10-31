package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archivist of Oghma — "Flash. Whenever an opponent searches their
// library, you gain 1 life and draw a card."
//
// AnOpponentSearchesTheirOwnLibrary (triggers_common.go) is the whole
// condition: EventSearchLibrary is emitted by the search path and,
// separately, by ShuffleLibrary with Label "shuffle" — only a real
// search should trigger, so the shuffle flavour is filtered out,
// otherwise every opponent's fetch-shuffle would double-trigger. And
// "their library" (CR 105's "their" is the searcher's own) is
// ev.Actor == ev.Target — #1335 gave the event a Target naming WHOSE
// library was searched, because Bribery-style search of another
// player's library (#1230) leaves those two apart: without this
// check, an opponent's Bribery search of MY library (Actor: them,
// Target: me) triggered Archivist just as their own tutor would,
// which is stronger than printed. Wan Shi Tong, Librarian shares the
// same condition (#1312, #1335).
func init() {
	Register(Spec{
		OracleID:        "08b13e1f-27ca-40a8-b5ed-88ac933d24bf",
		Name:            "Archivist of Oghma",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerMakesBlocker}, // #4734: a creature with flash is a blocker (sweep ruling 7)
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSearchLibrary, AnOpponentSearchesTheirOwnLibrary,
				"Archivist of Oghma — gain 1 life and draw", Do(
					// The draw waits for a prompt the life gain asks (CR 608.2c, #4155).
					GainLife{Amount: 1},
					DrawCards{N: 1},
				)),
		},
	})
}
