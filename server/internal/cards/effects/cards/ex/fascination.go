package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fascination — {X}{U}{U} Sorcery (Fate Reforged):
//
//	"Choose one —
//	 • Each player draws X cards.
//	 • Each player mills X cards."
//
// X is the announced value (CR 107.3). Each player acts in APNAP order; a
// library that holds fewer mills what it has (CR 701.17b), and a player who
// must draw from an empty library loses at the next state check (CR 704.5b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28f0765a-de09-4be3-b952-aee75b9ede03",
		Name:         "Fascination",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Each player draws X cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					for _, p := range effects.ApnapPlayers(ctx.Game) {
						if err := (effects.DrawCards{Player: p, N: ctx.X()}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ModeDoing("Each player mills X cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					for _, p := range effects.ApnapPlayers(ctx.Game) {
						if err := (effects.MillCards{Player: p, N: ctx.X()}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		),
	})
}
