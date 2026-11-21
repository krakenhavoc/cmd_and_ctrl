package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Conditioning — Sorcery {6}{B} (Scourge):
//
//	"Any number of target players each discard a number of cards equal to
//	 the greatest mana value among permanents you control."
//
// Any number of player targets, zero included (CR 601.2c), and one number
// read as the spell resolves (scgB1GreatestManaValue). Each still-legal
// target discards that many cards of their own choice; a player with fewer
// cards discards their hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37a04d5a-9272-42bd-b448-c658375bbbc3",
		Name:         "Cabal Conditioning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("any number of target players").WithCount(0, 0),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.ScgB1GreatestManaValue(ctx)
			if n <= 0 {
				return nil
			}
			for _, p := range effects.KozilekTargetPlayers(ctx) {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player: p,
					Source: item.SourceCardID,
					N:      n,
				})
			}
			return nil
		},
	})
}
