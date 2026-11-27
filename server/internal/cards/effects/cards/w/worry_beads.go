package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worry Beads — Artifact {3} (Mercadian Masques, S82 OLD batch 700, #5619):
//
//	"At the beginning of each player's upkeep, that player mills a card."
//
// One trigger for every upkeep, anyone's; the player whose upkeep it is, read
// off the event, mills one card (S82OldB700EachUpkeepThatPlayerMills).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab050d55-a47d-4880-b65a-29f562773d7a",
		Name:         "Worry Beads",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Worry Beads — that player mills a card", effects.S82OldB700EachUpkeepThatPlayerMills),
		},
	})
}
