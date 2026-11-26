package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaws of Stone — Sorcery {5}{R} (Shadowmoor, S82 OLD b530, #5591):
//
//	"Jaws of Stone deals X damage divided as you choose among any number of
//	 targets, where X is the number of Mountains you control as you cast this
//	 spell."
//
// Monstrous Onslaught's cast-time division with the Mountains you control as
// the amount: it is read once as the targets and shares are announced (CR
// 601.2d), so a Mountain lost in response changes nothing. Each target gets at
// least 1; one that is illegal on resolution takes nothing and its share is not
// reassigned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af6a8149-a849-4e6a-bf9d-e6b4bb9b1a04",
		Name:         "Jaws of Stone",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(0, 0).Dividing(effects.DivideBy(effects.OldB530DivideMountains)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
