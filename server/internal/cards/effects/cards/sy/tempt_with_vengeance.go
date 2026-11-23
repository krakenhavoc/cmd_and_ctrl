package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempt with Vengeance — Sorcery {X}{R} (Tempting offer):
//
//	"Tempting offer — Create X 1/1 red Elemental creature tokens with
//	 haste. Each opponent may create X 1/1 red Elemental creature tokens
//	 with haste. For each opponent who does, create X 1/1 red Elemental
//	 creature tokens with haste."
//
// X is the value announced for the spell; each token batch is created by
// the player the offer names.
func init() {
	tempt := effects.NewCmdr2S49Tempting("confirm/tempt-with-vengeance", "Tempt with Vengeance — create X 1/1 red Elementals with haste (the caster gets X more if you do)?", temptWithVengeanceDo)
	effects.Register(effects.Spec{
		OracleID:     "8e356df5-ca92-4be2-871e-8965c2510fbe",
		Name:         "Tempt with Vengeance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return tempt.Resolve(ctx)
		},
	})
}

func temptWithVengeanceDo(ctx *effects.Context, player uuid.UUID) error {
	if ctx.X() <= 0 {
		return nil
	}
	return effects.CreateToken{Controller: player, Template: effects.TokenCard("1/1 red Elemental with haste"), N: ctx.X()}.Apply(ctx)
}
