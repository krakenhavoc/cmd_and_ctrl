package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lupine Prototype — Artifact Creature — Wolf Construct {2}, 5/5 (Innistrad
// Remastered, #5948):
//
//	"This creature can't attack or block unless a player has no cards in hand."
//
// A layer-6 restriction on itself while every player still in the game holds at
// least one card. Hand sizes change without anything on the battlefield moving,
// so the static declares DependsOnHandSize and a draw, a discard or a play
// re-reads it. "A player" is any player, the Prototype's controller included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "303f91ae-1acb-4250-a4a5-71207c228dcc",
		Name:         "Lupine Prototype",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:             "This creature can't attack or block unless a player has no cards in hand.",
			Layer:             game.Layer6Ability,
			DependsOnHandSize: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && !gapB08APlayerHasNoCardsInHand(g)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
	})
}
