package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astral Slide — Enchantment {2}{W} (Onslaught):
//
//	"Whenever a player cycles a card, you may exile target creature. If
//	 you do, return that card to the battlefield under its owner's
//	 control at the beginning of the next end step."
//
// WheneverYouCycle's EventCycle watch with no condition on who cycled:
// any player's cycling, typecycling included, triggers it (CR 702.29f).
// The target is chosen as the trigger goes on the stack (CR 603.3d), and
// the "you may" is asked as it resolves (CR 603.5). The blink is Angel
// of Condemnation's delayed one (exileTargetsThenScheduleReturn): the
// card comes back under its owner's control as a new object at the next
// end step (CR 400.7, CR 603.7), and only a card that really reached
// exile is returned; a token exiled this way does not come back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70f14b10-8129-4249-a267-8b4be5d9befa",
		Name:         "Astral Slide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					On(game.EventCycle, AnyPlayer, "Astral Slide — exile target creature until the next end step",
						func(g *game.Game, item *game.StackItem) error {
							return exileTargetsThenScheduleReturn(NewContext(g, item), "Astral Slide — return the exiled card")
						}),
					"Astral Slide — exile target creature until the next end step?"),
				TargetCreature("target creature"),
			),
		},
	})
}
