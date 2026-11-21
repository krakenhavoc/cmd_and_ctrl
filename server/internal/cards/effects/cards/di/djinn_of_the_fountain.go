package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Djinn of the Fountain — Creature — Djinn {4}{U}{U}, 4/4 (Dominaria
// United, #5487):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, choose one —
//	 • This creature gets +1/+1 until end of turn.
//	 • Exile this creature. Return it to the battlefield under its owner's
//	 control at the beginning of the next end step.
//	 • Scry 1."
//
// A modal cast trigger (CR 603.3c). The pump is pinned to this permanent
// (CR 400.7). The blink exiles the source only while it is still the
// permanent the ability came from and schedules the delayed return at the
// next end step (Abuelo's exileCardsThenScheduleReturn); it comes back as a
// new object.
//
// No simplification.
func init() {
	const name = "Djinn of the Fountain"
	t := effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), name+" — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("This creature gets +1/+1 until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.ThisGetsUntilEndOfTurn(1, 1, name+" — +1/+1")(ctx.Game, item)
			}),
		effects.ModeDoing("Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.ExileThisThenScheduleReturn(ctx, item, name+" — return the exiled creature")
			}),
		effects.ModeWithPurpose(effects.ModeDoing("Scry 1.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.Scry{Player: item.Controller, N: 1}.Apply(ctx)
			}), game.Purpose{Scry: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:        "dcf28f91-1bf4-48be-a3d0-e061f443fd9e",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
