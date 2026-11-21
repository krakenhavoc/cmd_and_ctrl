package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunters' Feast — Sorcery {3}{G} (Magic 2011, #5579):
//
//	"Any number of target players each gain 6 life."
//
// Reward the Faithful's "any number of target players" at a fixed 6:
// zero players is a legal choice, and each player still a legal target
// as it resolves gains 6 (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df206fb3-2d98-439e-9f1a-161bdbdc47b2",
		Name:         "Hunters' Feast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("any number of target players").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, p := range effects.KozilekTargetPlayers(ctx) {
				if err := (effects.GainLife{Player: p, Amount: 6}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
