package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Atrocious Experiment — Sorcery {2}{B} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Target player mills two cards, draws two cards, and loses 2 life."
//
// The three instructions happen in order to the one target player (Brain
// Freeze's mill, then the draw, then the life loss, CR 608.2c): the draw is
// from the library the mill just changed, and a player who can't draw two
// loses at the next state-based check (CR 704.5b) and still loses the life.
// The player is read as the spell resolves; with it gone nothing happens
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f1c5040-ac46-4e8f-a5ff-8205e13a8893",
		Name:         "Atrocious Experiment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p, ok := ctx.ClauseTarget(0)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			if err := (effects.MillCards{Player: p.ID, N: 2}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.DrawCards{Player: p.ID, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p.ID, -2)
		},
	})
}
