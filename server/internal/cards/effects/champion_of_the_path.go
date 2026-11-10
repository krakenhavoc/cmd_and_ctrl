package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Champion of the Path — Creature — Elemental Sorcerer {3}{R}, 7/3
// (Lorwyn Eclipsed, #4458):
//
//	"As an additional cost to cast this spell, behold an Elemental and
//	 exile it. (Exile an Elemental you control or an Elemental card from
//	 your hand.)
//	 Whenever another Elemental you control enters, it deals damage equal
//	 to its power to each opponent.
//	 When this creature leaves the battlefield, return the exiled card to
//	 its owner's hand."
//
// The cost and the leave trigger are Champions of the Shoal's
// (BeholdAndExileCost, ReturnTheExiledCardToHand). The middle trigger
// fires for any other Elemental permanent entering under your control,
// read through the layered view (a changeling counts, and so does a
// kindred permanent, per the ruling). As it resolves, the entering
// object is read as it is now, or as it last existed on the battlefield
// once it has left (CR 608.2h, ctx.TriggeringPermanent): if it is not a
// creature then, it deals no damage (the ruling); otherwise it deals one
// damage instance equal to its power to each opponent (Origin of Thor's
// shape), the Elemental as the source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "b326ccd5-8ad2-441c-9d73-82ee6390dd36",
		Name:           "Champion of the Path",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdAndExileCost("an", "Elemental"),
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, championOfThePathAnotherElementalEntered,
				"Champion of the Path — it deals damage equal to its power to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || !slices.Contains(info.Characteristic.Types, "Creature") || info.Power <= 0 {
						return nil
					}
					return originOfThorDealToEachOpponent(ctx, item.Trigger.Event.CardID, info.Power)
				}),
			ReturnTheExiledCardToHand("Champion of the Path"),
		},
	})
}

// championOfThePathAnotherElementalEntered is "another Elemental you
// control enters". "Elemental" is a word of the card's text (CR 612.2).
func championOfThePathAnotherElementalEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Elemental"))
}
