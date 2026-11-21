package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skaab Wrangler — Creature — Human Wizard {1}{U}, 2/1 (Midnight Hunt, #5497):
//
//	"Tap three untapped creatures you control: Tap target creature."
//
// The cost is a TapOthers component, not the {T} symbol (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8de2ce5a-9345-4d2b-a248-868268f36836",
		Name:         "Skaab Wrangler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap three untapped creatures you control: Tap target creature.",
			Cost:    effects.S82MidB09TapThreeCreatures(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
