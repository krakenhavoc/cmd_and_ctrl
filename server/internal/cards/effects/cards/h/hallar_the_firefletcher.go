package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hallar, the Firefletcher — Legendary Creature — Elf Archer {1}{R}{G}, 3/3
// (Dominaria):
//
//	"Trample
//	 Whenever you cast a spell, if that spell was kicked, put a +1/+1 counter on
//	 Hallar, then Hallar deals damage equal to the number of +1/+1 counters on it
//	 to each opponent."
//
// A cast trigger with an intervening "if" (CR 603.4): the spell is yours and was
// kicked, read off the payment record the cast left on the stack item (a spell
// kicked more than once is kicked all the same). As it resolves the counter goes
// on first, then the count is read (counters from before and the new one;
// last-known if Hallar has left, CR 608.2h) for one damage instance to every
// opponent. Hallar's own cast does not trigger it: it is not on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "04559759-1afd-4785-badc-c3416041f016",
		Name:            "Hallar, the Firefletcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB209YouCastAKickedSpell,
				"Hallar, the Firefletcher — put a +1/+1 counter on Hallar, then it deals damage equal to the number of +1/+1 counters on it to each opponent",
				effects.OldB209HallarKicked),
		},
	})
}
