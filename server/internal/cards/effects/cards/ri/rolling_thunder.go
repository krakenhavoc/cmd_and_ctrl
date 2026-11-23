package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rolling Thunder — Sorcery {X}{R}{R} (#1658, unblocked by #1656):
//
//	"Rolling Thunder deals X damage divided as you choose among any
//	 number of targets."
//
// The card divide.go's own doc comment names as the canonical "any
// number of targets, amount is X" shape (Fire Covenant is the same
// shape with the amount paid in life instead of mana). X is paid in
// mana at cast (CR 601.2b) and IS the divided amount (CR 601.2d): "any
// number" is bounded above by X itself, since every target needs at
// least 1.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c47888b-28a5-4c43-9ee4-a9059e3c367d",
		Name:         "Rolling Thunder",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetAny().WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
