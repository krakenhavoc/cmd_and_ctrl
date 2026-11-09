package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alania's Pathmaker — Creature — Otter Wizard {2}{U}, 2/3 (Bloomburrow,
// #4467):
//
//	"When this creature enters, exile the top card of your library. Until
//	 the end of your next turn, you may play that card."
//
// Impulse draw for one card over a seat-turn window (ADR 0063), the
// Elemental Mascot way: the card may be played (a land too) until the end of
// its controller's next turn, paying its costs as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce36a450-68a6-4519-a9f0-0e3b772054eb",
		Name:         "Alania's Pathmaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Alania's Pathmaker — exile the top card of your library; until the end of your next turn, you may play it", Do(
				Instruction(func(ctx *Context) error {
					window := ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())
					return ExileTopWithPermission{
						From:    ctx.Controller(),
						GrantTo: ctx.Controller(),
						N:       1,
						Window:  &window,
					}.Apply(ctx)
				}),
			)),
		},
	})
}
