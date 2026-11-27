package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timesifter — Artifact {5} (Mirrodin, #5606):
//
//	"At the beginning of each upkeep, each player exiles the top card of their
//	 library. The player who exiled the card with the greatest mana value takes
//	 an extra turn after this one. If two or more players' cards are tied for
//	 greatest, the tied players repeat this process until the tie is broken."
//
// Triple Triad's exile chain (every player's library in APNAP order, no
// permission stamped) and then a judge: the owner of the card with the greatest
// mana value takes an extra turn (CR 500.7); with a tie the tied players run
// the chain again, the others sitting it out. A player with an empty library
// exiles nothing and cannot have the greatest card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99f8805f-6618-4838-a4e7-a3660e3b087f",
		Name:         "Timesifter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Timesifter — each player exiles the top card of their library; the player with the greatest mana value takes an extra turn after this one",
				effects.OldB607TimesifterUpkeep),
		},
	})
}
