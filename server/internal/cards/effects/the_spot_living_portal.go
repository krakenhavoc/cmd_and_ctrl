package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Spot, Living Portal — Legendary Creature — Human Scientist Villain
// {3}{W}{B}, 4/4 (Marvel's Spider-Man, #4460):
//
//	"When The Spot enters, exile up to one target nonland permanent and up
//	 to one target nonland permanent card from a graveyard.
//	 When The Spot dies, put him on the bottom of his owner's library. If you
//	 do, return the exiled cards to their owners' hands."
//
// The first trigger has two target clauses, each "up to one", and exiles
// every chosen object linked to The Spot (ExileTarget.Linked), so "the exiled
// cards" of the second trigger are exactly the cards still in exile that this
// object's first trigger exiled (a card that has since left exile lost the
// link, CR 400.7). The second trigger reads that link from The Spot's
// last-known object before moving him: he goes to the bottom only if he is
// still the same object in his owner's graveyard (a replacement that sent him
// elsewhere means he never died, and one that stopped the move means "if you
// do" is false); the cards come back to their owners' hands only if he
// really reached the library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5dfa7ad-9637-4da8-9ed2-0ac863617d9a",
		Name:         "The Spot, Living Portal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("The Spot — exile up to one target nonland permanent and up to one target nonland permanent card from a graveyard",
					theSpotExile),
				Clauses(
					TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1),
					TargetCardInGraveyard("up to one target nonland permanent card from a graveyard",
						And(Nonland(), Permanent())).WithCount(0, 1),
				)),
			WhenThisDies("The Spot — put him on the bottom of his owner's library; if you do, return the exiled cards to their owners' hands",
				theSpotBottomAndReturn),
		},
	})
}

// theSpotExile exiles each chosen object, linked to The Spot.
func theSpotExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for slot := 0; slot < 2; slot++ {
		for _, t := range ctx.ClauseTargets(slot) {
			if t.Kind != game.TargetCard || t.ID == uuid.Nil {
				continue
			}
			if err := (ExileTarget{Target: t.ID, Linked: true}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// theSpotBottomAndReturn is the dies trigger: the move to the library
// bottom, then (only if it happened) the exiled cards back to their owners.
func theSpotBottomAndReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	if err := (PutIntoLibrary{Card: item.SourceCardID, ToBottom: true}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	for _, id := range ids {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
