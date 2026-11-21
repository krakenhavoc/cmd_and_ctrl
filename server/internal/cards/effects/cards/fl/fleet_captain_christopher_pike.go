package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleet Captain Christopher Pike — Legendary Creature — Human Officer
// {2}{U}, 0/6 (Star Trek Commander, #4974):
//
//	"Yes — {1}, {T}: Untap two other target creatures.
//	 No — {2}, {T}: Tap two other target creatures."
//
// "Yes" and "No" are ability words (CR 207.2c). Each ability needs exactly
// two targets, neither of them Pike ("other": Another). A target that
// became illegal is skipped and the other is still tapped or untapped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0a9a7a2-c996-4e3e-817d-9283be690c39",
		Name:         "Fleet Captain Christopher Pike",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Yes — {1}, {T}: Untap two other target creatures.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.Another(effects.TargetCreature("two other target creatures").WithCount(2, 2)),
				Effect:  effects.UntapEachLegalTarget,
			},
			{
				Label:   "No — {2}, {T}: Tap two other target creatures.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.Another(effects.TargetCreature("two other target creatures").WithCount(2, 2)),
				Effect:  effects.TapEachLegalTarget,
			},
		},
	})
}
