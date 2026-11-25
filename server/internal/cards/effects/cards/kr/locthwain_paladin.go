package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Locthwain Paladin — Creature — Human Knight {3}{B}, 3/2 (Throne of
// Eldraine, #5516):
//
//	"Menace
//	 Adamant — If at least three black mana was spent to cast this spell,
//	 this creature enters with a +1/+1 counter on it."
//
// A CR 614.1c entry clause read off the spell's payment: the counter is
// put on as it enters when at least three black mana was actually spent to
// cast it (so Doubling Season sees it), and none when it was put onto the
// battlefield some other way. The menace is a printed keyword. WantsColors asks the auto-payer to
// concentrate black mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "6fea3c95-9663-4857-983e-eaf522efa1a8",
		Name:                       "Locthwain Paladin",
		Completeness:               effects.CompletenessFull,
		ReadsManaSpent:             game.ReadsOwnPayment,
		WantsColors:                []string{"B"},
		PrintedKeywords:            []string{"menace"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.OldB140CountersIfAdamantSpent(game.CounterPlusOne, "B")},
	})
}
