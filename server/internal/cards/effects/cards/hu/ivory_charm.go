package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivory Charm — Instant {W} (Mirage, S82 OLD b797, #5633):
//
//	"Choose one —
//	 • All creatures get -2/-0 until end of turn.
//	 • Tap target creature.
//	 • Prevent the next 1 damage that would be dealt to any target this turn."
//
// Three bullets, each with its own target clause (none for the first): the
// -2/-0 is Biting Rain's all-creatures boost, the tap Air Servant's body and
// the shield Bandage's one-point prevention (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dded2874-e80f-401a-b700-52c54dd6da60",
		Name:         "Ivory Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("All creatures get -2/-0 until end of turn."), game.Purpose{Answers: game.AnswerRestrict}),
			effects.ModeWithPurpose(effects.Mode("Tap target creature.", effects.TargetCreature("target creature")), game.Purpose{Answers: game.AnswerRestrict}),
			effects.ModeWithPurpose(effects.Mode("Prevent the next 1 damage that would be dealt to any target this turn.", effects.TargetAny()), game.Purpose{Answers: game.AnswerPrevent}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Label: "Ivory Charm — -2/-0 until end of turn"}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.TapTheFirstLegalTarget(ctx.Game, item)
			case ctx.HasMode(2):
				return effects.PreventNextOneDamageToTheTarget("Ivory Charm")(ctx.Game, item)
			}
			return nil
		},
	})
}
