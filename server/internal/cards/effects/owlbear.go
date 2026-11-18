package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Owlbear — Creature — Bird Bear {3}{G}{G}, 4/4 (Adventures in the Forgotten Realms):
//
//	Trample
//	 Keen Senses — When this creature enters, draw a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0fab815f-bccf-40ef-95e9-ba64b137fb8c",
		Name:            "Owlbear",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Owlbear — Keen Senses: draw a card", Do(DrawCards{N: 1})),
		},
	})
}
