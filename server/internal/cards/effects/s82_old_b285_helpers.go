package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b285_helpers.go — shared pieces of S82 older-set batch 285
// (Shadows over Innistrad: graveyard recursion, #5543).

// S82OldB285ReturnThisDiscardingTwo is the row "<cost>, Discard two cards:
// Return this card from your graveyard to the battlefield tapped." that
// Stitchwing Skaab and Ghoulsteed print (Advanced Stitchwing's ability at
// other prices): a graveyard activation (CR 113.6) whose discard is a cost.
func S82OldB285ReturnThisDiscardingTwo(manaCost string) ActivatedAbility {
	return ActivatedAbility{
		Label:   manaCost + ", Discard two cards: Return this card from your graveyard to the battlefield tapped.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    Plus(ManaCost(manaCost), DiscardN(2, "two cards")),
		Zones:   []game.ZoneKind{game.ZoneGraveyard},
		Effect:  ReturnThisFromGraveyardTapped,
	}
}

// S82OldB285FlyingToFirstTarget is "target creature gains flying until end
// of turn" as a trigger body: the first chosen target still legal as the
// trigger resolves gains the keyword (CR 608.2b).
func S82OldB285FlyingToFirstTarget(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: name + " — gains flying"}.Apply(ctx)
		}
		return nil
	}
}

// S82OldB285SameNameInYourGraveyard is the TargetsFrom clause "target card
// with the same name as that spell from your graveyard" on a cast trigger:
// the spell is the event's card, and the target is a card in the
// controller's graveyard that shares a name with it (CR 201.2a).
func S82OldB285SameNameInYourGraveyard(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	spell, ok := g.LookupCardForEffect(tc.Event.CardID)
	if !ok {
		return nil
	}
	names := ExtractionNames(spell)
	if len(names) == 0 {
		return nil
	}
	same := SharesANameWith(names)
	owner := source.Controller
	return &game.TargetSpec{
		Mode: "card", Label: "target card with the same name as that spell from your graveyard",
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.Owner == owner && same(c)
		},
		Min: 1, Max: 1,
	}
}

// S82OldB285HarnessOffer is the trigger's body: the chosen target, if it is
// still in the graveyard as the ability resolves, is offered as a paid cast
// (Sproutback Trudge's permission: the card's own mana cost, the other costs
// as printed). Declining leaves it in the graveyard.
func S82OldB285HarnessOffer(name string) Effect {
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
				Player:      item.Controller,
				Zone:        game.ZoneGraveyard,
				Duration:    g.UntilEndOfTurnDuration(),
				Cost:        c.ManaCost,
				AltCostKey:  game.AltCostKeyGrantedCost,
				Timing:      game.TimingFlash,
				CastOnly:    true,
				LapseOnPass: game.LapseStaysInHand,
				Source:      item.SourceCardID,
				SourceName:  name,
				Label:       name + " — cast " + c.Name + " from your graveyard",
			}, []game.Card{c})
			OfferResolutionCastOver(g, item.Controller, item.SourceCardID, name, game.MayCastKeywordPaidCast, DeclineLeaveInGraveyard, c.InstanceID)
			return nil
		}
		return nil
	}
}
