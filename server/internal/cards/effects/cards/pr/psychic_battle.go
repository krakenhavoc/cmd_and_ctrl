package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Battle — Enchantment {3}{U}{U} (Invasion):
//
//	"Whenever a player chooses one or more targets, each player reveals the
//	 top card of their library. The player who reveals the card with the
//	 greatest mana value may change the target or targets. If two or more
//	 cards are tied for greatest, the target or targets remain unchanged.
//	 Changing targets this way doesn't trigger abilities of permanents named
//	 Psychic Battle."
//
// #6228's seam: the EventTargetsChosen trigger (one per choosing, at an
// announcement or a change of targets) excepting a change made by an object
// named Psychic Battle, and the top-card contest that may change all the
// targets. Full: a tie, an empty library or a spell gone from the stack
// changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79f0d763-5a71-470a-867f-86fc72e9db6d",
		Name:         "Psychic Battle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAPlayerChoosesTargets("Psychic Battle",
				"Psychic Battle — each player reveals the top card of their library; the player with the greatest mana value may change the target or targets",
				effects.TopCardContestMayRetarget),
		},
	})
}
