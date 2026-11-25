package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ardenvale Paladin — Creature — Human Knight {3}{W}, 2/5 (Throne of
// Eldraine, #5516):
//
//	"Adamant — If at least three white mana was spent to cast this spell,
//	 this creature enters with a +1/+1 counter on it."
//
// A CR 614.1c entry clause read off the spell's payment: the counter is
// put on as it enters when at least three white mana was actually spent to
// cast it (so Doubling Season sees it), and none when it was put onto the
// battlefield some other way. WantsColors asks the auto-payer to
// concentrate white mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "30393cb0-f461-4899-8aae-609bc8378f36",
		Name:                       "Ardenvale Paladin",
		Completeness:               effects.CompletenessFull,
		ReadsManaSpent:             game.ReadsOwnPayment,
		WantsColors:                []string{"W"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.OldB140CountersIfAdamantSpent(game.CounterPlusOne, "W")},
	})
}
