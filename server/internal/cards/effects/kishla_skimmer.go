package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kishla Skimmer — Creature — Bird Scout {G}{U}, 2/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Whenever a card leaves your graveyard during your turn, draw a card.
//	 This ability triggers only once each turn."
//
// Attuned Hunter's trigger (left_graveyard.go): every route a card leaves
// your graveyard by counts, and cards that leave on an opponent's turn do
// nothing. "Triggers only once each turn" is TriggersOnlyOnceEachTurn,
// over OncePerBatch so several cards leaving as one event still make one
// trigger (Baron Bertram Graywater's shape). A trigger that was countered
// still spent the turn's use.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3275d80-112b-4839-a014-e44ef48d2e6b",
		Name:            "Kishla Skimmer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyardDuringYourTurn,
				"Kishla Skimmer — draw a card",
				Do(DrawCards{N: 1})))),
		},
	})
}
