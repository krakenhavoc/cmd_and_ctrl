package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// job_select.go — "Job select" (Final Fantasy): "When this Equipment
// enters, create a 1/1 colorless Hero creature token, then attach this to
// it." Living weapon with a Hero instead of a Germ, so it is the same
// trigger over the same shared body (createTokenAndAttachSource), not an
// engine seam. If the Equipment has left the battlefield in response, or
// is a new object by the time the trigger resolves, nothing is attached
// and the Hero stays unequipped.

// jobSelectHeroToken is the token every job-select Equipment makes.
const jobSelectHeroToken = "1/1 colorless Hero"

// JobSelect is the job select keyword ability for `cardName`.
func JobSelect(cardName string) game.TriggeredAbility {
	return WhenThisEnters(cardName+" — job select: create a 1/1 Hero and attach this to it",
		func(g *game.Game, item *game.StackItem) error {
			return createTokenAndAttachSource(g, item, jobSelectHeroToken)
		})
}

// AttachedIsAlso is "Equipped creature is a <subtype> in addition to its
// other types": a layer-4 addition to the creature types, not a
// replacement (compare SetAttachedTypes).
func AttachedIsAlso(subtype string) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Attached creature is a " + subtype + " in addition to its other types.",
		Layer:     game.Layer4Type,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !eotHasType(c.Subtypes, subtype) {
				c.Subtypes = append(c.Subtypes, subtype)
			}
		},
	}
}
