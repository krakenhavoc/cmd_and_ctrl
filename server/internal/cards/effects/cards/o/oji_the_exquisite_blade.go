package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oji, the Exquisite Blade — Legendary Creature — Human Monk {2}{W}{U}, 2/3
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"When Oji enters, you gain 2 life and scry 2.
//	 Whenever you cast your second spell each turn, exile up to one target
//	 creature you control, then return it to the battlefield under its owner's
//	 control."
//
// The enters trigger is two steps (gain, then scry). The second spell is Monk
// of the Open Hand's clock; the blink is Conjurer's Closet's, returning the
// card under its owner's control (FlickerFirstLegalTargetUnderOwner), and the
// target is picked as the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c4c556c-b5e3-493a-a7b0-38c55b7809ac",
		Name:         "Oji, the Exquisite Blade",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oji, the Exquisite Blade — you gain 2 life and scry 2",
				effects.Do(effects.GainLife{Amount: 2}, effects.Scry{N: 2})),
			effects.Targeting(
				effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
					"Oji, the Exquisite Blade — exile up to one target creature you control, then return it to the battlefield under its owner's control",
					effects.FlickerFirstLegalTargetUnderOwner),
				effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1)),
		},
	})
}
