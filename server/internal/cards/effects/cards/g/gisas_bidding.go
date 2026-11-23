package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gisa's Bidding — Sorcery {2}{B}{B} (Shadows over Innistrad, #5543):
//
//	"Create two 2/2 black Zombie creature tokens.
//	 Madness {2}{B} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Moan of the Unhallowed's two Zombies, without the flashback. Madness is
// grown by the engine from the one string (CR 702.35a, #657).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0abc993b-18c7-4bb8-aa58-0d06a41cb48f",
		Name:         "Gisa's Bidding",
		Completeness: effects.CompletenessFull,
		Madness:      "{2}{B}",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 2}),
	})
}
