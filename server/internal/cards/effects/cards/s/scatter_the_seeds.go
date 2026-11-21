package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scatter the Seeds — Instant {3}{G}{G} (Starter Commander Decks):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Create three 1/1 green Saproling creature tokens."
//
// Mob's convoke tap cost over Spore Swarm's three Saprolings.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c17d5153-cc74-4fbb-9d75-f7ee7b557b0a",
		Name:         "Scatter the Seeds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3, Answers: game.AnswerMakesBlocker},
		TapCost:      effects.Convoke(),
		OnResolve: effects.ResolveSteps(effects.CreateToken{
			Template: effects.TokenCard("1/1 green Saproling"),
			N:        3,
		}),
	})
}
