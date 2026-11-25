package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Embereth Paladin — Creature — Human Knight {3}{R}, 4/1 (Throne of
// Eldraine, #5516):
//
//	"Haste
//	 Adamant — If at least three red mana was spent to cast this spell,
//	 this creature enters with a +1/+1 counter on it."
//
// A CR 614.1c entry clause read off the spell's payment: the counter is
// put on as it enters when at least three red mana was actually spent to
// cast it (so Doubling Season sees it), and none when it was put onto the
// battlefield some other way. The haste is a printed keyword. WantsColors asks the auto-payer to
// concentrate red mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "eec31fe0-a633-4424-84d2-90fe88016310",
		Name:                       "Embereth Paladin",
		Completeness:               effects.CompletenessFull,
		ReadsManaSpent:             game.ReadsOwnPayment,
		WantsColors:                []string{"R"},
		PrintedKeywords:            []string{"haste"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.OldB140CountersIfAdamantSpent(game.CounterPlusOne, "R")},
	})
}
