package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyclops Tyrant — Creature — Cyclops {5}{R}, 3/4 (Magic 2014, #5563):
//
//	"Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
//	 This creature can't block creatures with power 2 or less."
//
// Intimidate, and a block rule on its own side of the pair: it may block only
// an attacker whose power is 3 or more, read as the block is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40ac0cb0-bdf7-447d-9875-0eb3cc5d0adc",
		Name:            "Cyclops Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(And(Creature(), PowerLE(2))),
				"it can't block creatures with power 2 or less"),
		},
	})
}
