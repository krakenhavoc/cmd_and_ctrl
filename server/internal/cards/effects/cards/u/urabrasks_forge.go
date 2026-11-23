package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urabrask's Forge — Artifact {2}{R}:
//
//	"At the beginning of combat on your turn, put an oil counter on this
//	 artifact, then create an X/1 red Phyrexian Horror creature token with
//	 trample and haste, where X is the number of oil counters on this
//	 artifact. Sacrifice that token at the beginning of the next end step."
//
// X is read after the counter lands (so a counter doubler counts), from the
// Forge's last-known information if it has left (CR 608.2h). The token is
// sized when it is made; a CR 603.7 delayed trigger sacrifices exactly that
// token at the next end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c4b2203-a937-4d91-8e39-2283fe686b45",
		Name:         "Urabrask's Forge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Urabrask's Forge — put an oil counter on it, then create an X/1 Phyrexian Horror",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.OnBattlefield(g, item.SourceCardID) && !effects.SourceIsNewObject(g, item) {
						if err := (effects.AddCounter{Target: item.SourceCardID, Kind: effects.S82OneB12Oil, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					x := effects.S82OneB12SourceOil(ctx)
					tmpl := effects.S82OneB12Token("Phyrexian Horror", "Token Creature — Phyrexian Horror", x, 1,
						[]string{"R"}, []string{"trample", "haste"})
					made, err := g.CreateTokensForEffect(item.Controller, ctx.TextToken(tmpl), 1, game.TokenEntryOptions{})
					if err != nil {
						return err
					}
					return effects.S82OneB12SacrificeAtEnd(ctx, "Urabrask's Forge — sacrifice the Horror", made)
				}),
		},
	})
}
