package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Restore — Sorcery {1}{G} (Commander 2013, S82 OLD b387):
//
//	"Put target land card from a graveyard onto the battlefield under your
//	 control."
//
// Reanimate's single-target body over a land card in any graveyard: the card
// enters untapped under the caster's control (CR 400.3: it stays its owner's),
// the target re-checked as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf3165bf-1d59-4e63-8650-bbda119e6335",
		Name:         "Restore",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target land card from a graveyard", effects.Land()),
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
