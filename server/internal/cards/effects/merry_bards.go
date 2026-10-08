package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merry Bards — Creature — Human Bard {2}{R}, 3/2:
//
//	"When this creature enters, you may pay {1}. When you do, create a
//	 Young Hero Role token attached to target creature you control. (If
//	 you control another Role on it, put that one into the graveyard.
//	 Enchanted creature has "Whenever this creature attacks, if its
//	 toughness is 3 or less, put a +1/+1 counter on it.")"
//
// "When you do" is a CR 603.12 reflexive trigger, so the target is
// chosen only after the {1} is paid. The Bards may target themselves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "699fa8e6-84f5-4010-a8fe-7b2e4f280ef9",
		Name:         "Merry Bards",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Merry Bards — you may pay {1}", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Merry Bards — pay {1} to create a Young Hero Role token attached to a creature you control?",
					OnPay: func(ctx *Context) error {
						return WhenYouDo("Merry Bards — create a Young Hero Role token attached to target creature you control",
							merryBardsYoungHeroBody).Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
