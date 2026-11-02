package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "af2275ea-1085-4145-aa6b-36924d53d88a",
		Name:         "School Daze",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Do Homework — Draw three cards."),
			Mode("Fight Crime — Counter target spell. Draw a card.", TargetSpell("target spell")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				return DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
			}
			if !ctx.HasMode(1) || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard ||
				!ctx.IsTargetLegal(item.Targets[0]) || ctx.Game.StackItemForEffect(item.Targets[0].ID) == nil {
				return nil
			}
			if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
