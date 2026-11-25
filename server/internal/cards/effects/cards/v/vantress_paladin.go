package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vantress Paladin — Creature — Human Knight {3}{U}, 2/2 (Throne of
// Eldraine, #5516):
//
//	"Flying
//	 Adamant — If at least three blue mana was spent to cast this spell,
//	 this creature enters with a +1/+1 counter on it."
//
// A CR 614.1c entry clause read off the spell's payment: the counter is
// put on as it enters when at least three blue mana was actually spent to
// cast it (so Doubling Season sees it), and none when it was put onto the
// battlefield some other way. The flying is a printed keyword. WantsColors asks the auto-payer to
// concentrate blue mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "1965ca0e-1357-4437-82d3-1a78353ab5f1",
		Name:                       "Vantress Paladin",
		Completeness:               effects.CompletenessFull,
		ReadsManaSpent:             game.ReadsOwnPayment,
		WantsColors:                []string{"U"},
		PrintedKeywords:            []string{"flying"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.OldB140CountersIfAdamantSpent(game.CounterPlusOne, "U")},
	})
}
