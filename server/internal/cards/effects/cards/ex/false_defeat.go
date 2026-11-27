package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// False Defeat — Sorcery {3}{W} (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Return target creature card from your graveyard to the battlefield."
//
// Breath of Life's text word for word: the card returns under the
// caster's control, and an illegal target does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b119579-38e1-4509-9dcb-019d5a160ac2",
		Name:         "False Defeat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
