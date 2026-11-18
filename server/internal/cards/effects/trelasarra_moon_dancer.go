package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trelasarra, Moon Dancer — Legendary Creature — Elf Cleric {G}{W}, 2/2
// (Adventures in the Forgotten Realms, S82):
//
//	"Whenever you gain life, put a +1/+1 counter on Trelasarra and scry 1."
//
// Ajani's Pridemate's trigger with Scry 1 after the counter, in the printed
// order. Once per life-gain event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd10ebc9-9f6c-4da2-9a86-263902002ebc",
		Name:         "Trelasarra, Moon Dancer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1}, // the scry half of the life-gain trigger
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				WheneverYouGainLife("Trelasarra, Moon Dancer — put a +1/+1 counter on it and scry 1",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, Scry{N: 1})),
				game.Purpose{LifeGainPayoff: &game.LifeGainPayoff{Counters: 1}}),
		},
	})
}
