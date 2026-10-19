package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seize the Spoils — Sorcery {2}{R}:
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards and create a Treasure token."
//
// Big Score's shape with one Treasure instead of two. Written out
// rather than reusing drawAndTreasure (whose `cards` parameter is
// always 2 across its existing callers) so that helper's signature
// doesn't have to grow a case only this card needs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "58f83528-9110-4895-b5ea-51b90af30a8d",
		Name:           "Seize the Spoils",
		Purpose:        game.Purpose{Draws: 2, Tokens: 1},
		Completeness:   CompletenessFull,
		AdditionalCost: DiscardCost(1),
		OnResolve: ResolveSteps(
			// The Treasure waits for a prompt the draws ask (CR 608.2c, #4155).
			DrawCards{N: 2},
			CreateToken{Template: TreasureToken(), N: 1},
		),
	})
}
