package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sabotender — Creature — Plant {1}{R}, 2/1 (EDHREC rank 2408):
//
//	"Reach
//	 Landfall — Whenever a land you control enters, this creature
//	 deals 1 damage to each opponent."
//
// The landfall pinger. Tireless Provisioner's condition (a land
// entered under the controller's control — played, fetched, or
// returned) with Impact Tremors' payoff; the Sabotender itself is the
// damage source, so a doubler sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65385311-1158-4e94-892a-683997706ca8",
		Name:            "Sabotender",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Sabotender — 1 damage to each opponent", game.LandfallPayoff{DamageEachOpponent: 1}, func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			}),
		},
	})
}
