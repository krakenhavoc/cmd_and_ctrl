package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garenbrig Paladin — Creature — Giant Knight {4}{G}, 4/4 (Throne of Eldraine,
// #5516):
//
//	"Adamant — If at least three green mana was spent to cast this spell,
//	 this creature enters with a +1/+1 counter on it.
//	 This creature can't be blocked by creatures with power 2 or less."
//
// Vantress Paladin's adamant entry counter, with Beloved Princess's block
// rule over the blocker's power (read live as the block is declared,
// CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "33fdbff6-0e89-4f02-b429-9b17869134cd",
		Name:                       "Garenbrig Paladin",
		Completeness:               effects.CompletenessFull,
		ReadsManaSpent:             game.ReadsOwnPayment,
		WantsColors:                []string{"G"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.OldB140CountersIfAdamantSpent(game.CounterPlusOne, "G")},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
