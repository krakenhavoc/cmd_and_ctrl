package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Crew — Sorcery {3}{R}:
//
//	"Create X 2/1 red Dwarf Berserker creature tokens, where X is the number of
//	 Vehicles you control plus the number of Equipment you control. For each of
//	 those tokens, you may attach an Equipment you control to it."
//
// X is counted as the spell resolves (CR 608.2h), an artifact that is both a
// Vehicle and an Equipment counting twice. The tokens are made through the
// keyed token door, so the attach sentence is told the ones that really entered
// (CR 608.2c): for each, in the order made, the controller picks any Equipment
// they control or none and it is attached (CR 701.3); an Equipment may move from
// one token to the next, as each instruction is its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ced584a8-7548-4b4e-8ccf-04b66d255539",
		Name:         "Reckless Crew",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.OldB072ReckEquipmentAndVehicles(ctx)
			if n <= 0 {
				return nil
			}
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: item.Controller,
				Source:     item.SourceCardID,
				Groups: []game.TokenGroup{{
					Template: ctx.TextToken(effects.TokenCard("2/1 red Dwarf Berserker")),
					Count:    n,
				}},
			}, ctx.ItemTail(effects.OldB072TokensMade))
		},
	})
}
