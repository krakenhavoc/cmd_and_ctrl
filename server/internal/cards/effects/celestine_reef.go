package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Celestine Reef — Plane — Luvion (#4870):
//
//	"Creatures without flying or islandwalk can't attack.
//	 Whenever chaos ensues, until a player planeswalks, you can't lose the
//	 game and your opponents can't win the game."
//
// The static is a layer-6 "can't attack" restriction on every creature that
// has neither flying nor islandwalk, read as layer 6 leaves the creature (a
// creature that gains flying may attack). The chaos ability grants the
// planar controller Platinum Angel's two game-end gates until the plane stops
// being face up (UntilYouPlaneswalkAway): a planeswalk ends them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2979597d-c3cc-4585-a9b7-7fa15641b030",
		Name:         "Celestine Reef",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Creatures without flying or islandwalk can't attack.",
			AppliesTo: func(target *game.Card, g *game.Game, _ *game.Card) bool {
				return target.IsCreature() &&
					WithoutKeyword("flying")(g, uuid.Nil, *target) &&
					WithoutKeyword("islandwalk")(g, uuid.Nil, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Celestine Reef — until a player planeswalks, you can't lose the game and your opponents can't win the game",
				func(g *game.Game, item *game.StackItem) error {
					d, ok := UntilYouPlaneswalkAway(g, item.SourceCardID)
					if !ok {
						return nil
					}
					for _, gate := range YouCantLoseOpponentsCantWin() {
						g.GrantGameEndGateForEffect(item.Controller, gate, "Celestine Reef", item.SourceCardID, d)
					}
					return nil
				}),
		},
	})
}
