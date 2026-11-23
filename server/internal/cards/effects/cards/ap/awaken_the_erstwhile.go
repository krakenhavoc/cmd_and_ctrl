package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "06fa6719-aa4f-4857-8029-e42d01232645",
		Name:         "Awaken the Erstwhile",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			players := effects.TablePlayers(ctx)
			counts := make([]int, len(players))
			for i, id := range players {
				n, err := effects.DiscardWholeHand(ctx.Game, id)
				if err != nil {
					return err
				}
				counts[i] = n
			}
			for i, id := range players {
				if counts[i] == 0 {
					continue
				}
				if err := (effects.CreateToken{Controller: id, Template: effects.BlackZombieToken(), N: counts[i]}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
