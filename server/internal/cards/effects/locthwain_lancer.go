package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Locthwain Lancer — Creature — Human Knight {4}{B}, 5/5 (Commander Legends
// of March of the Machine):
//
//	"Menace
//	 Whenever a nontoken Knight you control dies, each opponent loses 1 life
//	 and you draw a card."
//
// The Lancer sees its own death (CR 603.10a), read from last-known
// information.
func init() {
	Register(Spec{
		OracleID:        "a9baa117-ab06-4af5-9a09-4be423d7cbbd",
		Name:            "Locthwain Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, s82Mom09NontokenKnightYouControlDied,
				"Locthwain Lancer — each opponent loses 1 life and you draw a card",
				Do(EachOpponentLosesLife{N: 1}, DrawCards{N: 1})),
		},
	})
}
