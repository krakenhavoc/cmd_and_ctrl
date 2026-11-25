package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Experimental Overload — Sorcery {2}{U}{R} (Core Set 2021, #5511):
//
//	"Create an X/X blue and red Weird creature token, where X is the number of
//	 instant and sorcery cards in your graveyard. Then you may return an
//	 instant or sorcery card from your graveyard to your hand. Exile
//	 Experimental Overload."
//
// In printed order: X is counted as the spell resolves (it is on the stack, so
// it never counts itself), then the optional return from the graveyard as it
// then is, then the spell exiles itself (Temporal Mastery), so it never
// reaches the graveyard and cannot be returned by its own text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7851e4c8-020c-4a3d-88ed-0c97609cdfed",
		Name:         "Experimental Overload",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				x := len(effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() }))
				return effects.S82OldB110WeirdToken(ctx, x)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Experimental Overload — you may return an instant or sorcery card from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() },
					Min:      0,
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
