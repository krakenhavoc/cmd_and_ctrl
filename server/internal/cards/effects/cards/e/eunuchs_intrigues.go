package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eunuchs' Intrigues — Sorcery {2}{R}
// (Portal Three Kingdoms, #5622):
//
//	"Target opponent chooses a creature they control. Other creatures they control can't block this turn."
//
// The chosen creature is the only one of that opponent's creatures that can
// block; the rest are barred, read live, so a creature that arrives later
// is barred too (CR 611.2c). With no creature to choose, all of them are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e34f8aae-bf55-4bba-9524-c841591bb14b",
		Name:         "Eunuchs' Intrigues",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.FirstLegalTargetPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			return effects.OldB711ChooseACreatureThen(ctx, player, "Eunuchs' Intrigues — choose a creature you control; your other creatures can't block", effects.OldB711OthersCantBlock,
				func() error { return effects.OldB711NoCreatureCantBlock(ctx, player) })
		},
	})
}
