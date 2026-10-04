package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Awaken the Erstwhile — Sorcery {3}{B}{B}:
//
//	"Each player discards all the cards in their hand, then creates
//	 that many 2/2 black Zombie creature tokens."
//
// Windfall's shape with a token payoff: every hand goes first, each
// discard emitting its own event so discard payoffs queue while the
// spell resolves, and only then does each player create as many
// Zombies as THEY discarded (not the most anyone discarded). A player
// with an empty hand makes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06fa6719-aa4f-4857-8029-e42d01232645",
		Name:         "Awaken the Erstwhile",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			players := tablePlayers(ctx)
			counts := make([]int, len(players))
			for i, id := range players {
				n, err := discardWholeHand(ctx.Game, id)
				if err != nil {
					return err
				}
				counts[i] = n
			}
			for i, id := range players {
				if counts[i] == 0 {
					continue
				}
				if err := (CreateToken{Controller: id, Template: BlackZombieToken(), N: counts[i]}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
