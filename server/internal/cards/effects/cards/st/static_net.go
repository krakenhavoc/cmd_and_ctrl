package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Static Net — Enchantment {3}{W} (The Brothers' War):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield.
//	 When this enchantment enters, you gain 2 life and create a tapped
//	 Powerstone token."
//
// Two printed triggers, so two registered rows. The first is Banishing
// Light's (CR 610.3): the exile records its own return. The second is
// independent of it; the Powerstone enters tapped because the effect
// says so.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f285781-ee03-4a58-a85e-a33f5efffc9d",
		Name:         "Static Net",
		Completeness: effects.CompletenessFull,
		Triggered: append(effects.ExileUntilThisLeavesTriggers("Static Net"),
			effects.WhenThisEnters("Static Net — you gain 2 life and create a tapped Powerstone token",
				effects.Do(
					effects.GainLife{Amount: 2},
					effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1},
				))),
	})
}
