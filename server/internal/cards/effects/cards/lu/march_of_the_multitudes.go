package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// March of the Multitudes — Instant {X}{G}{W}{W} (EDHREC rank 3032):
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of
//	 that creature's color.)
//	 Create X 1/1 white Soldier creature tokens with lifelink."
//
// Selesnya's instant-speed army. Convoke is the S22 tap-permanents
// cost on the cast; X rides the cast and is read back with ctx.X();
// the tokens are b22WhiteSoldierLifelinkToken (Dawn of Hope's — the
// same printed token). Zero tokens for X=0 is legal and pointless,
// as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c26ab0d-80f6-4e5b-9d0e-af17c1519583",
		Name:         "March of the Multitudes",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Soldier with lifelink"), N: ctx.X()}.Apply(ctx)
		},
	})
}
