package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "40ac0cb0-bdf7-447d-9875-0eb3cc5d0adc",
		Name:            "Cyclops Tyrant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.PowerLE(2))),
				"it can't block creatures with power 2 or less"),
		},
	})
}
