package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ichor Slick — Sorcery {2}{B} (Dominaria Remastered, #5948):
//
//	"Target creature gets -3/-3 until end of turn.
//	 Cycling {2}
//	 Madness {3}{B}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a36fc2e-bb02-4222-a335-4c80e380e690",
		Name:         "Ichor Slick",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		Madness:      "{3}{B}",
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BoostTargetUntilEOT("Ichor Slick — -3/-3 until end of turn", -3, -3)(ctx.Game, item)
		},
	})
}
