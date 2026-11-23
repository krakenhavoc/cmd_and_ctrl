package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eldritch Pact — Sorcery {6}{B} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Target player draws X cards and loses X life, where X is the number of
//	 cards in their graveyard."
//
// X is the target player's graveyard as the spell resolves, counted once
// (CR 608.2h) before either instruction runs, so the draw does not change the
// life lost. Zero cards is zero of both.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "281ea9b6-055a-4274-90b3-c74311a6f863",
		Name:         "Eldritch Pact",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			p := ctx.Game.PlayerByIDForEffect(t.ID)
			if p == nil || p.Graveyard == nil {
				return nil
			}
			x := p.Graveyard.Size()
			if x <= 0 {
				return nil
			}
			if err := (effects.DrawCards{Player: t.ID, N: x}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -x)
		},
	})
}
