package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sleep —
//
// "Tap all creatures target player controls. Those creatures don't untap
// during that player's next untap step."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b93ff69-f195-4d72-8e1d-574c3e53bca8",
		Name:         "Sleep",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			ids := effects.B751CreatureIDs(ctx.Game, ts[0].ID)
			return (effects.TapAndFreeze{Targets: ids, Player: ts[0].ID, Label: "Sleep"}).Apply(ctx)
		},
	})
}
