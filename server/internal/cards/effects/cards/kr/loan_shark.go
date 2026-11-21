package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loan Shark — Creature — Shark Rogue {3}{U}, 3/4 (Outlaws of Thunder Junction):
//
//	"When this creature enters, if you've cast two or more spells this turn,
//	 draw a card.
//	 Plot {3}{U}"
//
// An intervening "if" (CR 603.4): read as the creature enters and again as
// the trigger resolves (CondSpellsCast, the cast tally of the turn). A
// plotted Loan Shark cast for free on a later turn is itself a spell cast,
// so it counts toward the two. Plot is the ordinary special action
// (CR 702.170).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "15030c1d-7c0a-4273-a022-eb98d6c7a1ce",
		Name:           "Loan Shark",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 1},
		SpecialActions: []game.SpecialAction{effects.Plot("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondSpellsCast{AtLeast: 2},
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.YouHaveCastSpellsThisTurn(2)),
					"Loan Shark — draw a card (cast two or more spells this turn)",
					effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
