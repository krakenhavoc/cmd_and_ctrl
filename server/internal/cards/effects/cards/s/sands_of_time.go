package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sands of Time — Artifact {4} (Visions, S82 older sets batch 792, #5631):
//
//	"Each player skips their untap step.
//	 At the beginning of each player's upkeep, that player simultaneously untaps
//	 each tapped artifact, creature, and land they control and taps each untapped
//	 artifact, creature, and land they control."
//
// Stasis's skip (CR 500.11) over every seat, and a trigger for each player's
// upkeep, the upkeep's player read off the event as the trigger resolves. The two
// sets (tapped, untapped) are fixed before either changes, and a permanent that is
// more than one of the three types is flipped once. Tapped for mana or not, an
// untapped permanent is tapped, so a permanent that untaps one upkeep is tapped
// the next.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8f69b48-a3b2-4324-852d-11fad19d6432",
		Name:         "Sands of Time",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Old792PlayersSkipUntapSteps("Sands of Time: skip untap step")},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Sands of Time — that player untaps each tapped artifact, creature, and land they control and taps each untapped one", effects.Old792SandsOfTimeFlip),
		},
	})
}
