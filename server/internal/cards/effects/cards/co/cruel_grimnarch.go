package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Grimnarch — Creature — Phyrexian Cleric {5}{B}, 5/5:
//
//	"Deathtouch
//	 When this creature enters, each opponent discards a card. For each
//	 opponent who can't, you gain 4 life."
//
// One run: every opponent with a card chooses their discard, and once the
// last has landed you gain 4 life for each opponent who discarded none. An
// opponent with no card in hand can't discard (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c27b7cd-dcbf-45a7-9e15-a6d8dfd3a087",
		Name:            "Cruel Grimnarch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cruel Grimnarch — each opponent discards a card; for each who can't, you gain 4 life", effects.CruelGrimnarchEnters),
		},
	})
}
