package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tendrils of Despair — Sorcery {B} (Weatherlight, #5629):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Target opponent discards two cards."
//
// The sacrifice is paid as the spell is cast (CR 601.2f). The opponent
// chooses the two cards; a smaller hand gives up what it has
// (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "3b00f415-2496-4333-be89-601fd837fe7c",
		Name:           "Tendrils of Despair",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetOpponentDiscards(2)(ctx.Game, item)
		},
	})
}
