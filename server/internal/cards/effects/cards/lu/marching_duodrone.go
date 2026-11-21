package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marching Duodrone — Artifact Creature — Construct {2}, 2/2 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Whenever this creature attacks, each player creates a Treasure token."
//
// Every player still in the game gets a Treasure under their own control,
// in turn order starting from the active player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "939b77f1-d7c4-4e52-b487-5866145ace01",
		Name:         "Marching Duodrone",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Marching Duodrone — each player creates a Treasure token", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.ApnapPlayers(g) {
					if id == uuid.Nil {
						continue
					}
					if err := (effects.CreateToken{Controller: id, Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
