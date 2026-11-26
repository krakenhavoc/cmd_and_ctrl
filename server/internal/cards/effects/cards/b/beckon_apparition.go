package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beckon Apparition — Instant {W/B} (Gatecrash, #5565):
//
//	"Exile target card from a graveyard. Create a 1/1 white and black Spirit
//	 creature token with flying."
//
// Rotten Reunion's exile-then-token with a required target: the card must be
// in a graveyard as the spell resolves, or the whole spell does nothing
// (CR 608.2b). The token is made whether or not the exile did anything else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10762505-ebaa-47de-9f3e-08d48f2ac240",
		Name:         "Beckon Apparition",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTargetCards(ctx.Game, ctx.Item) }),
			effects.CreateToken{Template: effects.TokenCard("1/1 white and black Spirit with flying"), N: 1},
		),
	})
}
