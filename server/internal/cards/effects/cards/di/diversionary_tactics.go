package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diversionary Tactics — Enchantment {3}{W} (Apocalypse, S82 OLD, #5614):
//
//	"Tap two untapped creatures you control: Tap target creature."
//
// Minister of Impediments's tap row paid by tapping two other creatures. That
// cost is not {T}, so summoning sickness does not stop a creature from paying it
// (CR 302.6); a target gone by resolution is not tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "880367e4-50b9-49d0-b596-d462d9b42855",
		Name:         "Diversionary Tactics",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two untapped creatures you control: Tap target creature.",
			Cost:    effects.OldB629TapTwoUntappedCreaturesCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
