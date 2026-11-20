package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bnsB13OtherCreatureYouControlEnters is the entry condition "another creature
// under the source's controller's control enters the battlefield".
func bnsB13OtherCreatureYouControlEnters(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && entering.IsCreature() && entering.Controller == src.Controller
}

// bnsB13OtherCreaturesEnterWithPowerCountersAsMutants is Master Biomancer's
// replacement: each other creature its controller controls enters with
// additional +1/+1 counters equal to the Biomancer's power (as it is when
// the creature enters, counters and anthems included) and as a Mutant in
// addition to its other types (CR 614.1c, CR 205.3d).
func BnsB13OtherCreaturesEnterWithPowerCountersAsMutants(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: bnsB13OtherCreatureYouControlEnters,
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, src.CurrentPower())
			ev.AddSubtypeAtETB("Mutant", label)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// bnsB13FreeCastFromGraveyard is "you may cast target <card> from <a>
// graveyard without paying its mana cost" as a reflexive body's step: the
// still-legal target card is offered a free cast during the resolution
// (CR 608.2g, #4250), with no exile-instead rider. The cast is a grant over
// the card in its graveyard ({0}, so X is locked at 0 and additional costs
// are still paid); a decline, or the holder's next priority pass, leaves the
// card where it is.
func BnsB13FreeCastFromGraveyard(sourceName string) func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	return func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
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
				Player:      item.Controller,
				Zone:        game.ZoneGraveyard,
				Duration:    g.UntilEndOfTurnDuration(),
				Cost:        "{0}",
				AltCostKey:  game.AltCostKeyGrantedCost,
				Timing:      game.TimingFlash,
				CastOnly:    true,
				LapseOnPass: game.LapseStaysInHand,
				Source:      item.SourceCardID,
				SourceName:  sourceName,
				Label:       sourceName + " — cast " + c.Name + " without paying its mana cost",
			}, []game.Card{c})
			OfferResolutionCastOver(g, item.Controller, item.SourceCardID, sourceName, game.MayCastKeywordFreeCast, DeclineLeaveInGraveyard, c.InstanceID)
			return nil
		}
		return nil
	}
}

// returnTargetInstantOrSorceryFromYourGraveyard is "return target instant or
// sorcery card from your graveyard to your hand" for a trigger whose target
// clause is TargetCardInGraveyard(... YouOwn(), Or(Instant(), Sorcery())):
// the target is re-checked as the ability resolves (CR 608.2b), and a card
// that has left the graveyard is not returned (Pinnacle Monk, Living Lightning).
func ReturnTargetInstantOrSorceryFromYourGraveyard(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	if !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}.Apply(ctx)
}
