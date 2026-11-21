package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nessian Wanderer — Creature — Satyr Scout {1}{G}, 1/3 (Commander
// Masters):
//
//	"Constellation — Whenever an enchantment you control enters, look at the
//	 top three cards of your library. You may reveal a land card from among
//	 them and put that card into your hand. Put the rest on the bottom of
//	 your library in a random order."
//
// Boughside Wanderers' look-and-take (LookAtTopThenMayTakeToHand) under a
// trigger on an enchantment of yours entering. The trigger is "an
// enchantment you control", so Nessian Wanderer, which is not an
// enchantment, never triggers it for itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45703364-660a-4e73-a467-5ee47d936c1a",
		Name:         "Nessian Wanderer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, gapB103EnchantmentEnteredUnderYourControl,
				"Nessian Wanderer — look at the top three cards; you may reveal a land card and put it into your hand",
				LookAtTopThenMayTakeToHand(3, Land(), 1,
					"Nessian Wanderer — reveal a land card and put it into your hand")),
		},
	})
}
