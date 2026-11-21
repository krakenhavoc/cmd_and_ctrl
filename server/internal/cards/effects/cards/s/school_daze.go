package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// School Daze — Instant {3}{U}{U} (Marvel's Spider-Man, #4460):
//
//	"Choose one —
//	 • Do Homework — Draw three cards.
//	 • Fight Crime — Counter target spell. Draw a card."
//
// Only the Fight Crime mode has a target. If that spell is gone or
// illegal when School Daze resolves, the mode does nothing at all, the
// draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af2275ea-1085-4145-aa6b-36924d53d88a",
		Name:         "School Daze",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Do Homework — Draw three cards."),
			effects.Mode("Fight Crime — Counter target spell. Draw a card.", effects.TargetSpell("target spell")),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				return effects.DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
			}
			if !ctx.HasMode(1) || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard ||
				!ctx.IsTargetLegal(item.Targets[0]) || ctx.Game.StackItemForEffect(item.Targets[0].ID) == nil {
				return nil
			}
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
