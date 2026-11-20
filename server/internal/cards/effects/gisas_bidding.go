package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0abc993b-18c7-4bb8-aa58-0d06a41cb48f",
		Name:         "Gisa's Bidding",
		Completeness: CompletenessFull,
		Madness:      "{2}{B}",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 2}),
	})
}
