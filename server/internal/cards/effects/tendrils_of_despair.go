package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "3b00f415-2496-4333-be89-601fd837fe7c",
		Name:           "Tendrils of Despair",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetOpponentDiscards(2)(ctx.Game, item)
		},
	})
}
