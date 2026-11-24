package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sproutback Trudge — Creature — Fungus Beast {7}{G}{G}, 9/7 (Commander
// 2021, #5504):
//
//	"This spell costs {X} less to cast, where X is the amount of life you
//	 gained this turn.
//	 Trample
//	 At the beginning of your end step, if you gained life this turn, you
//	 may cast this creature from your graveyard."
//
// The discount is a self cost modifier over the turn's life-gained tally
// (the turn's total, not the net change, CR 119.9), reducing generic mana
// only, and it applies from the graveyard too. The end-step trigger works
// from the graveyard (Zones) and is an intervening if (CR 603.4): checked
// as the step begins and again as it resolves. It offers the cast during
// the resolution (CR 608.2g): the card's own mana cost, with the discount,
// timing ignored, and declining leaves it in the graveyard.
//
// No simplification.
func init() {
	trigger := effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.B24YourEndStepAndYouGainedLifeThisTurn(ev, source, g)
	}, "Sproutback Trudge — you may cast this creature from your graveyard", sproutbackTrudgeOffer)
	trigger.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:        "d9826660-0f34-4d82-8915-cf2796ab8d9a",
		Name:            "Sproutback Trudge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(
				func(q game.CostQuery) int { return effects.B15LifeGainedThisTurn(q.Game, q.Controller) },
				"This spell costs {X} less to cast, where X is the amount of life you gained this turn."),
		},
		Triggered: []game.TriggeredAbility{trigger},
	})
}

// sproutbackTrudgeOffer is the trigger's body: the intervening if again,
// then the paid cast of this card from the graveyard, offered during the
// resolution.
func sproutbackTrudgeOffer(g *game.Game, item *game.StackItem) error {
	if effects.B15LifeGainedThisTurn(g, item.Controller) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	const name = "Sproutback Trudge"
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      item.Controller,
		Zone:        game.ZoneGraveyard,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        c.ManaCost,
		AltCostKey:  game.AltCostKeyGrantedCost,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInHand,
		Source:      c.InstanceID,
		SourceName:  name,
		Label:       name + " — cast it from your graveyard",
	}, []game.Card{c})
	effects.OfferResolutionCastOver(g, item.Controller, c.InstanceID, name, game.MayCastKeywordPaidCast, effects.DeclineLeaveInGraveyard, c.InstanceID)
	return nil
}
