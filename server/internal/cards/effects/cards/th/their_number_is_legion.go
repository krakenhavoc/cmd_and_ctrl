package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Their Number Is Legion — Sorcery {X}{B}{B}{B}{B} (Warhammer 40,000
// Commander, #5948):
//
//	"Create X tapped 2/2 black Necron Warrior artifact creature tokens,
//	 then you gain life equal to the number of artifacts you control.
//	 Exile Their Number Is Legion.
//	 You may cast this card from your graveyard."
//
// The tokens enter tapped and are artifacts, so they count toward the life
// gained, which is read after they are made. The spell exiles itself as the
// last step, so a cast from the graveyard (or from the hand) leaves it in
// exile, never back in the graveyard; only a countered or discarded copy
// reaches the graveyard, which is what the second ability is for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "8cd45064-2ad9-4247-902e-aefaf3bc1a43",
		Name:          "Their Number Is Legion",
		Completeness:  effects.CompletenessFull,
		XMatters:      true,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.X() <= 0 {
					return nil
				}
				return effects.CreateTokenAdvanced{
					Spec: effects.Token(effects.TokenCard("2/2 black Necron Warrior artifact")).EntersTapped(),
					N:    ctx.X(),
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsArtifact() })
				return effects.GainLife{Amount: n}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
