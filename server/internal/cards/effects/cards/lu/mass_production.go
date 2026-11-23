package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mass Production — Sorcery {5}{W} (The Brothers' War, #5482):
//
//	"Create four 1/1 colorless Soldier artifact creature tokens."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20150b58-a442-472c-8191-b44581322eeb",
		Name:         "Mass Production",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard(effects.BroB05SoldierToken), N: 4}.Apply(ctx)
		},
	})
}
