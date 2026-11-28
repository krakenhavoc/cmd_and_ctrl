package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower of Murmurs — Artifact {4} (Mirrodin, S82 OLD, #5606):
//
//	"{8}, {T}: Target player mills eight cards."
//
// Millstone's ability at eight cards: any player may be the target, the
// cards go from the top of their library to their graveyard as it
// resolves, and a library of fewer than eight mills what is there (CR
// 701.17a). A target that is no longer legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c383b28c-b319-49cd-acf1-4721a1301bbb",
		Name:         "Tower of Murmurs",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}, {T}: Target player mills eight cards.",
			Cost:    effects.Plus(effects.ManaCost("{8}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.B31MillChosenPlayer(8),
		}},
	})
}
