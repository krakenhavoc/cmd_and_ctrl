package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scheming Fence — Creature — Human Citizen {W}{U}, 2/3:
//
//	"As this creature enters, you may choose a nonland permanent.
//	 Activated abilities of the chosen permanent can't be activated.
//	 This creature has all activated abilities of the chosen permanent
//	 except for loyalty abilities. You may spend mana as though it were
//	 mana of any color to activate those abilities."
//
// The choice is CR 614.12's, stored on the Fence as the OBJECT chosen
// (game.Card.ChosenPermanent): a pick of at most one nonland permanent,
// "you may" being the empty pick. A chosen permanent that leaves the
// battlefield is gone for good, since it comes back as a new object.
//
// The other three halves are Drana and Linvala's shapes over that one
// permanent (ADR 0155):
//
//   - The gate judges the SOURCE of an activation, mana abilities
//     included, so the chosen permanent is stopped and the Fence's copy
//     of its ability is not (that copy's source is the Fence).
//   - The borrow lends the chosen permanent's finished rows except
//     loyalty abilities (BorrowExceptLoyalty, CR 606).
//   - The rider covers only the rows this borrow lent.
//
// No simplification.
func init() {
	const clause = "This creature has all activated abilities of the chosen permanent except for loyalty abilities."
	effects.Register(effects.Spec{
		OracleID:     "cef87a98-61a0-4a88-950f-867144f685e3",
		Name:         "Scheming Fence",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChoosePermanentAsEnters("Scheming Fence — you may choose a nonland permanent", effects.Nonland()),
		ActivationRestrictions: []game.ActivationRestriction{
			effects.ChosenPermanentCantActivate("Scheming Fence — activated abilities of the chosen permanent can't be activated."),
		},
		Static: []game.StaticAbility{effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Except:  game.BorrowExceptLoyalty,
			Lenders: effects.ChosenPermanentLender(),
		})},
		AnyColorSpend: effects.BorrowedAbilitiesSpendableAsAnyColor(
			"You may spend mana as though it were mana of any color to activate those abilities.", clause),
	})
}
