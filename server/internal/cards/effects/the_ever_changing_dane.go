package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ever-Changing 'Dane — Legendary Creature — Shapeshifter {W}{U}{B}, 3/3
// (Dominaria United Commander):
//
//	"{1}, Sacrifice another creature: The Ever-Changing 'Dane becomes a copy of
//	 the sacrificed creature, except it has this ability."
//
// Dimir Doppelganger's and Unstable Shapeshifter's duration copy with no
// stated duration (BecomeCopy, CopyIndefinite): it lasts until 'Dane leaves the
// battlefield or copies the next creature. The copy is of the creature "as it
// last existed on the battlefield" (CR 608.2h, CR 707.2): the sacrifice is a
// cost, so the creature is gone when the ability resolves and its copiable
// values come from the last-known record (a token copies as the token it was).
// "Except it has this ability" is a grant of the ability as a named bundle
// (CR 707.9a), so a 'Dane that has become a Bear can activate it again. A copy
// of a creature that is legendary is legendary: nothing says otherwise.
//
// No simplification.
const everChangingDaneGrant = "the-ever-changing-dane/this-ability"

var everChangingDaneAbility = ActivatedAbility{
	Label: "{1}, Sacrifice another creature: The Ever-Changing 'Dane becomes a copy of the sacrificed creature, except it has this ability.",
	Cost:  Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
	// A sacrifice outlet at will (the cost sacrifices another creature).
	Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		refs := ctx.Paid().SacrificedObjects
		info, ok := ctx.SacrificedPermanent()
		if !ok || len(refs) == 0 {
			return nil
		}
		copyOf := BecomeCopy{
			Targets:  []uuid.UUID{ctx.Source()},
			Of:       refs[0].ID,
			Duration: CopyIndefinite,
			Except:   func(v *game.PrintedValues) { v.GrantAbility(everChangingDaneGrant) },
			Label:    "The Ever-Changing 'Dane — becomes a copy of the sacrificed creature",
		}
		if info.Copiable != nil {
			last := info.Copiable.Clone()
			copyOf.Values = &last
		} else if card, found := g.LookupCardForEffect(refs[0].ID); found {
			own := game.OwnPrintedValues(card)
			copyOf.Values = &own
		} else {
			return nil
		}
		return copyOf.Apply(ctx)
	},
}

func init() {
	Register(Spec{
		OracleID:     "bda0468f-1c72-4d3d-ad94-15d446c75313",
		Name:         "The Ever-Changing 'Dane",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       everChangingDaneGrant,
			Text:      "{1}, Sacrifice another creature: This creature becomes a copy of the sacrificed creature, except it has this ability.",
			Activated: []ActivatedAbility{everChangingDaneAbility},
		}},
		Activated: []ActivatedAbility{everChangingDaneAbility},
	})
}
