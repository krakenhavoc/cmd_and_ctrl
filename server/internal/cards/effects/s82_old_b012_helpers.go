package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b012_helpers.go — shared pieces of the Secret Lair Drop batch 12
// cards (#5462).

// OldB12CastTargetInstantFromYourGraveyardExilingIt is "you may cast target
// <card> from your graveyard. If that spell would be put into a graveyard,
// exile it instead" with the card's mana cost PAID (Toshiro Umezawa). It is
// FreeCastTargetFromYourGraveyardExilingIt's twin without the {0}: a
// cast-only, flash-speed permission offered during the resolution
// (CR 608.2g, #4250), the spell exiled whether it resolves, is countered or
// fizzles (ExileInsteadOfGraveyard, #3863). The target is re-checked as the
// ability resolves (CR 608.2b).
func OldB12CastTargetInstantFromYourGraveyardExilingIt(sourceName string) Effect {
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
				AltCostKey:              game.AltCostKeyGrantedCost,
				Timing:                  game.TimingFlash,
				CastOnly:                true,
				ExileInsteadOfGraveyard: true,
				LapseOnPass:             game.LapseStaysInHand,
				Source:                  item.SourceCardID,
				SourceName:              sourceName,
				Label:                   sourceName + " — cast " + c.Name,
			}, []game.Card{c})
			OfferResolutionCastOver(g, item.Controller, item.SourceCardID, sourceName, game.MayCastKeywordPaidCast, DeclineLeaveInGraveyard, c.InstanceID)
			return nil
		}
		return nil
	}
}

// OldB12OpponentCreatureDied is "a creature an opponent controls dies": the
// permanent left the battlefield for a graveyard as a creature, under the
// control of a player other than the source's controller (last-known
// information, CR 603.10a), a token included.
func OldB12OpponentCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
		return false
	}
	c, _ := g.LookupCardForEffect(ev.CardID)
	if !LeftAsType(ev, c, "creature") {
		return false
	}
	controller := LeftUnderControlOf(ev, c)
	return controller != uuid.Nil && controller != source.Controller
}
