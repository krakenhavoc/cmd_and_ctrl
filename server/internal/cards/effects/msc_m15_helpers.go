package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// msc_m15_helpers.go — shared pieces of Marvel Super Heroes Commander, M–Z
// batch 15 (#4455). Append-only (the clone gate's rule).

// freeCastTargetFromYourGraveyardExilingIt is the resolution of "you may
// cast target <card> from your graveyard without paying its mana cost. If
// that spell would be put into your graveyard, exile it instead"
// (Torrential Gearhulk, Victor Timely). The cast is a grant, the way every
// "you may cast it" a resolution offers is (ADR 0099 §4): a free ({0}, so
// X is locked at 0 and additional costs are still paid), any-timing
// (CR 608.2g), cast-only permission over the target card in the graveyard,
// offered during the resolution (#4250). The holder's next priority pass
// closes it only as the fallback, and the card simply stays where it is.
// "If that spell would be put into your graveyard, exile it instead" is the
// permission's ExileInsteadOfGraveyard (#3863): the spell is exiled whether
// it resolves, is countered or fizzles. The target is re-checked as the
// ability resolves (CR 608.2b): one that left the graveyard is not cast.
func freeCastTargetFromYourGraveyardExilingIt(sourceName string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			z := g.FindCardZoneForEffect(t.ID)
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok || z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:                  item.Controller,
				Zone:                    game.ZoneGraveyard,
				Duration:                g.UntilEndOfTurnDuration(),
				Cost:                    "{0}",
				AltCostKey:              game.AltCostKeyGrantedCost,
				Timing:                  game.TimingFlash,
				CastOnly:                true,
				ExileInsteadOfGraveyard: true,
				// A pass closes the window; nothing moves, the card
				// stays in the graveyard.
				LapseOnPass: game.LapseStaysInHand,
				Source:      item.SourceCardID,
				SourceName:  sourceName,
				Label:       sourceName + " — cast " + c.Name + " without paying its mana cost",
			}, []game.Card{c})
			// #4250: cast during the resolution (CR 608.2g).
			offerResolutionCastOver(g, item.Controller, item.SourceCardID, sourceName, game.MayCastKeywordFreeCast, declineLeaveInGraveyard, c.InstanceID)
			return nil
		}
		return nil
	}
}

// destroyTheFirstLegalTarget is "destroy target <permanent>" for a clause
// with one target: the first target still legal as the ability resolves
// (CR 608.2b), or nothing (The Witch's Vanity's first chapter, The Spear of
// Bashenga).
func destroyTheFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := b16FirstLegalTargetCard(ctx); ok {
		return DestroyTarget{Target: id}.Apply(ctx)
	}
	return nil
}
