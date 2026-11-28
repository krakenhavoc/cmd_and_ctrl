package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b768_helpers.go — shared pieces of S82 OLD batch 768 (#5628):
// Tempest enters and dies triggers.

// PutThisDiedCardOnTopOfItsOwnersLibrary is the body of "When this creature
// dies, put it on top of its owner's library" (Dark Revenant, Avenging Angel):
// the card moves only if it is still in a graveyard as the same object that
// died (CR 400.7), so a card exiled in response is left alone.
func PutThisDiedCardOnTopOfItsOwnersLibrary(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !SameGraveyardObject(z, item) {
		return nil
	}
	return PutIntoLibrary{Card: item.SourceCardID}.Apply(NewContext(g, item))
}

// ReturnThisToHandUnlessYouWonTheFlip is "if you lose the flip, return this
// creature to its owner's hand" (Scoria Wurm, Wild Wurm): it goes back only if
// the call was lost and the source is still the object that flipped (CR 400.7).
func ReturnThisToHandUnlessYouWonTheFlip(ctx *Context, r game.CoinFlipped) error {
	if WonFirstFlip(r) || SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return BounceToHand{Target: ctx.Source()}.Apply(ctx)
}

// DracoplasmEntersSacrificing is Dracoplasm's "As this creature enters,
// sacrifice any number of creatures. This creature's power becomes the total
// power of those creatures and its toughness becomes their total toughness":
// Devour's entry choice (DevourOver), minus the counters, with a Then that
// reads each sacrificed creature's last-known power and toughness (CR 608.2h)
// and sets ReplacementEvent.EntersWithBasePT from the totals.
func DracoplasmEntersSacrificing(name string) game.ReplacementEffect {
	r := entersOnlyIf(name,
		name+" — sacrifice any number of creatures as it enters.",
		&game.EntryCardChoice{
			Action:    game.EntryCardSacrifice,
			Matches:   func(c game.Card) bool { return c.IsCreature() },
			AnyNumber: true,
			Then: func(g *game.Game, ev *game.ReplacementEvent, picked []uuid.UUID) error {
				power, toughness := 0, 0
				for _, id := range picked {
					if info, ok := g.LastKnownPermanentForEffect(id); ok {
						power += info.Power
						toughness += info.Toughness
					}
				}
				ev.EntersWithBasePT = true
				ev.EntersBasePower, ev.EntersBaseToughness = power, toughness
				return nil
			},
		})
	// Declining (sacrificing none) is not a graveyard trip: the creature enters
	// either way.
	r.Replace = nil
	r.RowLabel = "As this creature enters, sacrifice any number of creatures; its power and toughness become their totals."
	return r
}
