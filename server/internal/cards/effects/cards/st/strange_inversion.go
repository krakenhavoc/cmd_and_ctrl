package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strange Inversion — Instant — Arcane {2}{R}:
//
//	"Switch target creature's power and toughness until end of turn.
//	Splice onto Arcane {1}{R}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "540f8779-cf29-4582-8b7f-95ebcae92925",
		Name:         "Strange Inversion",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{1}{R}"),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SwitchEachTargetPT(ctx, "Strange Inversion — switch power and toughness")
		},
	})
}
