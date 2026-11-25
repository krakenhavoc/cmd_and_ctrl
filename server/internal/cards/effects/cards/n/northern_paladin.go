package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Northern Paladin — Creature — Human Knight {2}{W}{W}, 3/3 (Secret Lair
// Drop):
//
//	"{W}{W}, {T}: Destroy target black permanent."
//
// Target black permanent is any black permanent, a creature or not, anyone's,
// read through the layers (CR 105.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5975294-508a-453e-893a-2fbea2487d17",
		Name:         "Northern Paladin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}, {T}: Destroy target black permanent.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target black permanent", effects.OfColor("B")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
