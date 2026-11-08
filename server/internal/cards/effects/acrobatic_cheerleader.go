package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acrobatic Cheerleader — Creature — Human Survivor {1}{W}, 2/2
// (Duskmourn: House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, put a flying counter on it. This ability
//	 triggers only once."
//
// Defiant Survivor's Survival trigger (survival.go), whose "if" is an
// intervening if asked as the phase begins and again as the ability
// resolves, with a flying counter (a keyword counter, CR 122.1b).
//
// "This ability triggers only once" is TriggersOnlyOnce (#5245): the
// engine records the first trigger of this object and declines every
// later one, on any turn, even if the first was countered or the counter
// was removed (the 2024-09-20 rulings). An untapped Cheerleader does not
// trigger, so it does not spend the ability. One that leaves and returns
// is a new object and may trigger again (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8ce3f18-26fe-47e8-be8b-0d89d7de7038",
		Name:         "Acrobatic Cheerleader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnce(Survival("Acrobatic Cheerleader — put a flying counter on it",
				Do(CounterOnThis{Kind: game.CounterFlying, N: 1}))),
		},
	})
}
