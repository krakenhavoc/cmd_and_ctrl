package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aphetto Grifter — Creature — Human Wizard {2}{U}, 1/1 (Onslaught):
//
//	"Tap two untapped Wizards you control: Tap target permanent."
//
// Shieldmage Elder's tap-two cost over Wizards (tapTwoUntapped). The
// cost has no {T} symbol, so the Grifter may tap itself as one of the
// two and a Wizard that arrived this turn may pay it (CR 302.6). Any
// permanent is a legal target, a land included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9d7b7a9-5107-4720-9f92-b49e372372b8",
		Name:         "Aphetto Grifter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two untapped Wizards you control: Tap target permanent.",
			Cost:    effects.TapTwoUntapped("two untapped Wizards you control", "Wizard"),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
