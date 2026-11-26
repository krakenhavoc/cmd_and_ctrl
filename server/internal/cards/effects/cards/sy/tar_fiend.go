package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tar Fiend — {5}{B} Creature — Elemental 4/4 (Shards of Alara, S82 OLD b514):
//
//	"Devour 2 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with twice that many +1/+1 counters on it.)
//	 When this creature enters, target player discards a card for each creature
//	 it devoured."
//
// The devoured count is read as the trigger is put on the stack (CR 702.82b);
// the target player chooses the discards (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82bc2ee1-9b51-4b3b-804b-1b4e8aab9a4f",
		Name:         "Tar Fiend",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Devour("Tar Fiend", 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OldB514DevouredTrigger("Tar Fiend — target player discards a card for each creature it devoured",
				effects.OldB514TargetPlayerDiscardsItemAmount), effects.TargetPlayer("target player")),
		},
	})
}
