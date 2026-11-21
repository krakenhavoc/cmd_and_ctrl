package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Filigree Racer — Artifact — Vehicle {3}{R}, 5/5:
//
//	"When this Vehicle enters, you get {E}{E}{E}{E} (four energy
//	 counters).
//	 Whenever this Vehicle attacks, you may pay {E}{E}. When you do,
//	 target instant or sorcery card in your graveyard gains jump-start
//	 until end of turn. (You may cast that card from your graveyard by
//	 discarding a card in addition to paying its other costs. Then exile
//	 it.)
//	 Crew 1"
//
// The energy is Aetherstorm Roc's (WhenThisEntersYouGetEnergy and
// whenThisAttacksMayPayEnergy: the payment holds the declare attackers
// step). Paying it opens a reflexive trigger (CR 603.12) that targets an
// instant or sorcery card in the controller's graveyard as it goes on the
// stack, and grants that card Niv-Mizzet, Supreme's jump-start terms for
// the turn (CR 702.133a): its own mana cost plus a discard, exiled as it
// leaves the stack. The target is re-checked as the trigger resolves
// (CR 608.2b). Crew 1 is the shared crew row.
func init() {
	const name = "Filigree Racer"
	effects.Register(effects.Spec{
		OracleID:     "5d93b4a9-0d51-45bd-a0e7-90737d019822",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 4},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy(name, 4),
			effects.WhenThisAttacksMayPayEnergy(name, 2, "give an instant or sorcery card in your graveyard jump-start until end of turn", filigreeRacerPaid),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect(name),
		}},
	})
}

// filigreeRacerPaid opens the reflexive trigger once the energy is paid.
// An on-disk key.
var filigreeRacerPaid = effects.EnergyPaid("pay-unless/filigree-racer-jump-start", func(g *game.Game, item *game.StackItem) error {
	return effects.ReflexiveTrigger{
		Label: "Filigree Racer — target instant or sorcery card in your graveyard gains jump-start until end of turn",
		Body:  filigreeRacerJumpStartBody,
	}.Apply(effects.NewContext(g, item))
})

var filigreeRacerJumpStartBody = game.ReflexiveBody("filigree-racer/jump-start", effects.SimpleBody(filigreeRacerGrantJumpStart),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()))
	}))

// filigreeRacerGrantJumpStart grants the chosen card jump-start for the
// turn: a cast permission from the graveyard at the card's own mana cost
// with a discard in addition, exiled as it leaves the stack.
func filigreeRacerGrantJumpStart(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		card, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:                item.Controller,
			Zone:                  game.ZoneGraveyard,
			CastOnly:              true,
			AltCostKey:            game.AltCostKeyJumpStart,
			DiscardCardInAddition: true,
			ExileOnResolution:     true,
			Source:                item.SourceCardID,
			SourceName:            "Filigree Racer",
			Label:                 "Jump-start — discard a card (Filigree Racer)",
		}, []game.Card{card})
		return nil
	}
	return nil
}
