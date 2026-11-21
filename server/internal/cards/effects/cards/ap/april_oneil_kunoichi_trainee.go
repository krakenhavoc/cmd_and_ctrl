package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// April O'Neil, Kunoichi Trainee — {1}{W} Legendary Creature — Human Ninja
// 2/2 (Teenage Mutant Ninja Turtles, #4457):
//
//	"When April O'Neil enters, scry 2.
//	 April O'Neil can't be blocked by creatures with power 3 or greater."
//
// The scry is Faerie Seer's enters trigger (CR 701.22a). The evasion is
// Bilbo, Unexpected Adventurer's block rule, the blocker's power read live as
// blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f166376-2739-411c-8f42-96f2ef28ef86",
		Name:         "April O'Neil, Kunoichi Trainee",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("April O'Neil, Kunoichi Trainee — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerGE(3), "creatures with power 3 or greater"),
		},
	})
}
