package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Power Pack — Legendary Creature — Human Hero {1}{R}{G}{W}{U}, 4/4:
//
//	"Flying, vigilance, trample, haste
//	 Whenever Power Pack deals combat damage to a player, exile target
//	 instant or sorcery card from your graveyard chosen at random. At
//	 the beginning of your next upkeep, you may cast that card without
//	 paying its mana cost. If that spell would be put into your
//	 graveyard, exile it instead."
//
// The target is ChosenAtRandom (#1815): drawn uniformly from the
// instant and sorcery cards in your graveyard as the trigger goes on the
// stack, through the keyed random stream; with none the trigger is
// removed (CR 603.3d). If the card is exiled, a delayed trigger (CR
// 603.7) waits for your next upkeep and, as it resolves, grants the
// cast Torrential Gearhulk's does (ADR 0099 §4): free ({0}, so X is 0
// and additional costs are still paid), any timing (CR 608.2g),
// cast-only, offered during the trigger's resolution (#4250) and closed by
// your next pass only as the fallback, with "exile it instead" as the
// permission's ExileInsteadOfGraveyard. A card that has left exile by
// then cannot be cast this way.
//
// No simplification.
func init() {
	hit := WheneverThisDealsCombatDamageToAPlayer("Power Pack — exile target instant or sorcery card from your graveyard chosen at random",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ExileTarget{Target: id, ThenTail: withCarry(ctx.ItemTail(powerPackExiled), id)}.Apply(ctx)
		})
	hit.Targets = ChosenAtRandom(TargetCardInGraveyard("target instant or sorcery card from your graveyard chosen at random",
		Or(Instant(), Sorcery()), YouOwn()))
	Register(Spec{
		OracleID:        "3d5c98ff-fcf1-42bd-9533-2648791a4f45",
		Name:            "Power Pack",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "trample", "haste"},
		Triggered:       []game.TriggeredAbility{hit},
	})
}

// powerPackCastExiled is the delayed trigger's body: grant the free cast
// of the payload card while it is still in exile.
func powerPackCastExiled(g *game.Game, item *game.StackItem) error {
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		z := g.FindCardZoneForEffect(t.ID)
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:                  item.Controller,
			Zone:                    game.ZoneExile,
			Duration:                g.UntilEndOfTurnDuration(),
			Cost:                    "{0}",
			AltCostKey:              game.AltCostKeyGrantedCost,
			Timing:                  game.TimingFlash,
			CastOnly:                true,
			ExileInsteadOfGraveyard: true,
			LapseOnPass:             game.LapseStaysInExile,
			Source:                  item.SourceCardID,
			SourceName:              "Power Pack",
			Label:                   "Power Pack — cast " + c.Name + " without paying its mana cost",
		}, []game.Card{c})
		// #4250: cast as the delayed trigger resolves (CR 608.2g).
		offerResolutionCastOver(g, item.Controller, item.SourceCardID, "Power Pack",
			game.MayCastKeywordFreeCast, declineLeaveInExile, c.InstanceID)
	}
	return nil
}

// powerPackExiled is the clause after the exile (#4253 slice 9).
var powerPackExiled = CardTail("power-pack/after-exile", func(ctx *Context, o game.TailOutcome) error {
	id := tailCarry0(o)
	exiled := len(o.IDs) == 1

	if !exiled {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Power Pack — you may cast that card without paying its mana cost",
		Cards:              []uuid.UUID{id},
		Body:               powerPackCastExiledBody,
	}.Apply(ctx)
})
