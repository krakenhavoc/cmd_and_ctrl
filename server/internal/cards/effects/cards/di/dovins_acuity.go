package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dovin's Acuity — Enchantment {1}{W}{U} (Ravnica Allegiance):
//
//	"When this enchantment enters, you gain 2 life and draw a card.
//	 Whenever you cast an instant spell during your main phase, you may return
//	 this enchantment to its owner's hand."
//
// The cast trigger is checked as the spell is cast, on your own turn in a main
// phase; "you may" is asked as the trigger resolves, and the return does nothing
// if the enchantment has left or come back as a new object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "898e480d-c44e-4080-910a-f53e4ee471e1",
		Name:         "Dovin's Acuity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dovin's Acuity — you gain 2 life and draw a card",
				effects.Do(effects.GainLife{Amount: 2}, effects.DrawCards{N: 1})),
			effects.Optional(effects.DuringYourMainPhase(effects.WheneverYouCast(effects.Instant(),
				"Dovin's Acuity — you may return this enchantment to its owner's hand",
				effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)))),
				"Dovin's Acuity — return this enchantment to its owner's hand?"),
		},
	})
}
