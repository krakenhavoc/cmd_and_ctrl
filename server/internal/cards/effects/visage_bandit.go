package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visage Bandit — Creature — Shapeshifter Rogue {3}{U}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"You may have this creature enter as a copy of a creature you
//	 control, except it's a Shapeshifter Rogue in addition to its other
//	 types.
//	 Plot {2}{U} (You may pay {2}{U} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// Glasspool Mimic's clone with plot: the candidates are the controller's
// own creatures, not targeted (CR 115.10a), the copy is optional, and the
// exception adds Shapeshifter and Rogue to the copied subtypes (CR
// 707.9b). Declined, or with nothing to copy, it is the printed 2/2.
// Cast from plot it still enters from the stack, so the copy is offered
// all the same. Plot is the ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "db11596a-6704-427b-aaef-0211e7822127",
		Name:           "Visage Bandit",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{U}")},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Visage Bandit", aCreatureYouControl, alsoAShapeshifterRogue),
		},
	})
}
