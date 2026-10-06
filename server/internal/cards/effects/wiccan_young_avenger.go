package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wiccan, Young Avenger — Legendary Creature — Mutant Warlock Hero
// {3}{R}, 3/3:
//
//	"Whenever you cast a noncreature spell, exile the top card of your
//	 library. Until your next end step, you may play that card."
//
// The trigger goes on the stack above the spell that cast it, so the
// exiled card is playable in time to respond with an instant. The
// window is game.UntilYourNextEndStep (#2373).
func init() {
	Register(Spec{
		OracleID:     "a317d5f4-a998-4f72-b080-5345bc7cc668",
		Name:         "Wiccan, Young Avenger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(),
				"Wiccan, Young Avenger — exile the top card; play it until your next end step",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopNUntilYourNextEndStep(NewContext(g, item), 1)
				}),
		},
	})
}
