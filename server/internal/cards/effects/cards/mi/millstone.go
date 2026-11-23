package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Millstone — Artifact {2} (Antiquities, #5640):
//
//	"{2}, {T}: Target player mills two cards."
//
// Nephalia Drownyard's mill on an artifact, two cards for three: the
// player is picked at announce and re-checked at resolution, and either
// player is a legal target. A library with fewer than two cards mills
// what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3212e47a-5492-4c50-9d4a-6ea562f1a6e1",
		Name:         "Millstone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target player mills two cards.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.B31MillChosenPlayer(2),
		}},
	})
}
