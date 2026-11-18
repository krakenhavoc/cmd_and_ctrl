package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reality Heist — Instant {5}{U}{U} (Kamigawa: Neon Dynasty, #5493):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Look at the top seven cards of your library. You may reveal up to two
//	 artifact cards from among them and put them into your hand. Put the rest
//	 on the bottom of your library in a random order."
//
// Affinity is Assert Authority's self cost modifier; the dig is Adaptive
// Omnitool's (takeRevealRestBottomRandom) with a cap of two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "569da5d3-6af2-4021-a464-b30ebaaf6fa0",
		Name:         "Reality Heist",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
				Match:    Artifact(),
				Max:      2,
				Optional: true,
				Label:    "Reality Heist — reveal up to two artifact cards and put them into your hand",
				Key:      takeRevealRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
