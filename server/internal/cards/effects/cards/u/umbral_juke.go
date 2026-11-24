package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Umbral Juke — Instant {2}{B} (Strixhaven: School of Mages, S82 batch 047,
// #5502):
//
//	"Choose one —
//	 • Target player sacrifices a creature or planeswalker of their choice.
//	 • Create a 2/1 white and black Inkling creature token with flying."
//
// The first bullet is Deadly Brew's edict on one targeted player: they
// choose among their creatures and planeswalkers (CR 701.21), and with
// none they sacrifice nothing. The second is Inkling Summoning's token,
// the shared 2/1 white and black Inkling with flying row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d56afdf6-51e0-4151-a86c-4c4827a50a0e",
		Name:         "Umbral Juke",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target player sacrifices a creature or planeswalker of their choice.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
						effects.SacrificeSpec("a creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
						"Umbral Juke — sacrifice a creature or planeswalker")
					return nil
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Create a 2/1 white and black Inkling creature token with flying.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("2/1 white and black Inkling with flying"), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1}),
		),
	})
}
