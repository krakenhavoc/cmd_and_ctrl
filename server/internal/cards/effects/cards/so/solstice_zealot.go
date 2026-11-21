package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Solstice Zealot — Creature — Rhino Cleric {2}{W}, 2/3:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}: Tap target creature."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ec63280-ae5b-4892-8854-de2d8127ea83",
		Name:         "Solstice Zealot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Solstice Zealot", 2),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}: Tap target creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(1)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
