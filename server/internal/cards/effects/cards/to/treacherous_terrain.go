package cardsto

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treacherous Terrain — Sorcery {6}{R}{G}:
//
//	"Treacherous Terrain deals damage to each opponent equal to the number
//	 of lands that player controls.
//	 Basic landcycling {2}"
//
// Each opponent is dealt their own land count, read as the spell resolves
// (CR 120.3). Basic landcycling is the cycling family's BasicLandcycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eced76b8-b29a-4c89-b77a-cbae99f56cb7",
		Name:         "Treacherous Terrain",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82BnsB09DamageEachOpponentBy(ctx, func(opp uuid.UUID) int {
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == opp && c.IsLand() {
						n++
					}
				}
				return n
			})
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
