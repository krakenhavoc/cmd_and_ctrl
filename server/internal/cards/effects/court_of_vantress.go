package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Court of Vantress — Enchantment {2}{U}{U} (Wilds of Eldraine Commander,
// #5368):
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, choose up to one other target
//	 enchantment or artifact. If you're the monarch, you may create a token
//	 that's a copy of it. If you're not the monarch, you may have this
//	 enchantment become a copy of it, except it has this ability."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// checked again as it resolves (CR 608.2b). Which half applies is read as
// it resolves — "if you're the monarch" is a clause of the effect — and the
// "you may" is asked then (CR 608.2d). The token is an ordinary token copy
// (CR 707.2). Becoming a copy has no stated duration (CR 611.2a); "has this
// ability" is a named grant in the copiable values (CR 707.9a), so the Court
// keeps working as the thing it copied, the way Thespian's Stage does.
//
// No simplification.
const courtOfVantressGrant = "court-of-vantress/this-ability"

var courtOfVantressTrigger = Optional(Targeting(
	AtYourUpkeep("Court of Vantress — copy up to one other target enchantment or artifact",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			targets := ctx.LegalTargets()
			if len(targets) == 0 || targets[0].Kind != game.TargetCard {
				return nil
			}
			if YoureTheMonarch(g, item.Controller) {
				return CreateTokenCopy{Controller: ctx.Controller(), Copy: targets[0].ID, N: 1}.Apply(ctx)
			}
			return BecomeCopy{
				Targets:  []uuid.UUID{ctx.Source()},
				Of:       targets[0].ID,
				Duration: CopyIndefinite,
				Except:   func(v *game.PrintedValues) { v.GrantAbility(courtOfVantressGrant) },
				Label:    "Court of Vantress — becomes a copy",
			}.Apply(ctx)
		}),
	Another(TargetPermanent("up to one other target enchantment or artifact", Or(Enchantment(), Artifact())).WithCount(0, 1))),
	"Court of Vantress — copy it?")

func init() {
	Register(Spec{
		OracleID:     "ad82b43b-b15e-48f4-86ab-5ee2c448cf14",
		Name:         "Court of Vantress",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       courtOfVantressGrant,
			Text:      "At the beginning of your upkeep, choose up to one other target enchantment or artifact. If you're the monarch, you may create a token that's a copy of it. If you're not the monarch, you may have this enchantment become a copy of it, except it has this ability.",
			Triggered: []game.TriggeredAbility{courtOfVantressTrigger},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Vantress"),
			courtOfVantressTrigger,
		},
	})
}
