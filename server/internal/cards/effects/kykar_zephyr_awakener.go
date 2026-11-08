package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kykar, Zephyr Awakener — Legendary Creature — Bird Wizard {2}{W}{U},
// 3/4 (Foundations, #4465):
//
//	"Flying
//	 Whenever you cast a noncreature spell, choose one —
//	 • Exile another target creature you control. Return that card to the
//	   battlefield under its owner's control at the beginning of the next
//	   end step.
//	 • Create a 1/1 white Spirit creature token with flying."
//
// The Vision's noncreature cast trigger with a plain choose-one. The first
// bullet's target rides its own mode (Charming Prince's shape) and is
// another creature you control (Another, CR 109.1); its body is the shared
// flicker (exileCardsThenScheduleReturn): one exile, then a CR 603.7
// delayed trigger that returns the card under its owner's control at the
// next end step, as a new object. The second bullet is the Spirit from the
// token table.
//
// No simplification.
func init() {
	t := WheneverYouCast(Noncreature(), "Kykar, Zephyr Awakener — you cast a noncreature spell: choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Exile another target creature you control. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
			Another(TargetCreature("another target creature you control", YouControl())),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				ref, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return exileCardsThenScheduleReturn(ctx, []uuid.UUID{ref.ID}, "Kykar, Zephyr Awakener — return the exiled card")
			}),
		ModeDoing("Create a 1/1 white Spirit creature token with flying.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "0156c94b-e64c-49de-a46c-a203c5e05e98",
		Name:            "Kykar, Zephyr Awakener",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
