package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sadistic Augermage — Creature — Human Wizard {2}{B}, 3/1 (Ravnica: City of
// Guilds, S82 OLD batch 576, #5600):
//
//	"When this creature dies, each player puts a card from their hand on top of
//	 their library."
//
// Each player, in turn order from the active player (CR 101.4), picks a card
// from their own hand and puts it on top of their own library; a player with
// one card puts that one, and a player with none puts none. The prompts go to
// the players at once and each places only its own player's card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f08c592-8ddd-4f52-95c3-5649cbb24d3a",
		Name:         "Sadistic Augermage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Sadistic Augermage — each player puts a card from their hand on top of their library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, p := range effects.ApnapPlayers(g) {
						if err := (effects.PutFromHandOnTopInAnyOrder{Player: p, N: 1,
							Label: "Sadistic Augermage — put a card from your hand on top of your library"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
