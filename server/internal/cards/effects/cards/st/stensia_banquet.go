package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stensia Banquet — Sorcery {2}{R} (Eldritch Moon, S82 OLD, #5541):
//
//	"Stensia Banquet deals damage to target opponent or planeswalker equal
//	 to the number of Vampires you control.
//	 Draw a card."
//
// Midnight Arsonist's Vampire count (read through the layered subtypes, a
// changeling counts), taken as the spell resolves; the draw happens even if
// the target has become illegal and nothing is dealt (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "089e7086-3f8e-4b10-a660-fcf4fc99604e",
		Name:         "Stensia Banquet",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetOpponentOrPlaneswalker(),
		Purpose:      stensiaBanquetPurpose(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				x := effects.VowB05VampiresControlledBy(ctx.Game, ctx.Controller(), "Vampire")
				return effects.SourceDealsDamageToEachLegalTarget(x)(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}

func stensiaBanquetPurpose() game.Purpose {
	p := effects.ForTargets(effects.DamageToTarget(0, 1))
	p.Draws = 1
	return p
}
