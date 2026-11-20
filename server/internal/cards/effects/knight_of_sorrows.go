package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of Sorrows — Creature — Human Knight {4}{W}, 3/3 (Ravnica
// Allegiance, #5521):
//
//	"This creature can block an additional creature each combat.
//	 Afterlife 1 (When this creature dies, create a 1/1 white and black
//	 Spirit creature token with flying.)"
//
// Foriysian Brigade's extra block (CanBlockAdditional, CR 509.1b).
// Afterlife is the dies trigger it reminds (CR 702.135a), Syndicate
// Messenger's 1/1 white and black Spirit with flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50f644d6-2597-4ad6-8ea8-6b09056f90c0",
		Name:         "Knight of Sorrows",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CanBlockAdditional(selfOnly, 1)},
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Knight of Sorrows — afterlife 1: create a 1/1 white and black Spirit creature token with flying",
				knightOfTheLastBreathSpirits(1)),
		},
	})
}
