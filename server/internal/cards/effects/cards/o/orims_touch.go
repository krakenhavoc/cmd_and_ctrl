package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orim's Touch — Instant {W} (Invasion, S82 OLD #5616):
//
//	"Kicker {1} (You may pay an additional {1} as you cast this spell.)
//	 Prevent the next 2 damage that would be dealt to any target this turn. If
//	 this spell was kicked, prevent the next 4 damage that would be dealt to
//	 that permanent or player this turn instead."
//
// Bandage's one-use shield on the target until end of turn (CR 615.7), its
// size read off the cast. The "instead" replaces the 2, so a kicked cast
// shields 4 and never 6. A target that has gone by resolution gets no shield
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "cc22c210-efd0-494c-8560-448b038b3c5f",
		Name:          "Orim's Touch",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerPrevent},
		Targets:       effects.TargetAny(),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 2
			if ctx.WasKicked() {
				n = 4
			}
			return effects.PreventNextDamageToTheTarget("Orim's Touch", n)(ctx.Game, item)
		},
	})
}
