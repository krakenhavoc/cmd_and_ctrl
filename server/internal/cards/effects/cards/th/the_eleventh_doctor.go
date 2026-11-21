package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Eleventh Doctor — Legendary Creature — Time Lord Doctor {1}{W}{U}, 3/2
// (Doctor Who, #5948):
//
//	"I. AM. TALKING! — Whenever The Eleventh Doctor deals combat damage to a
//	 player, you may exile a card from your hand with a number of time counters
//	 on it equal to its mana value. If it doesn't have suspend, it gains
//	 suspend.
//	 {2}: Target creature with power 3 or less can't be blocked this turn."
//
// The trigger asks for one card from the controller's hand (or none: "you
// may"), exiles it, and puts time counters on it equal to its mana value
// (X is 0 in a hand, CR 107.3b), and a card without suspend is marked as
// having been given it (Game.GainSuspendForEffect, as Sinister Concierge and
// Kang Prime), so it counts down in the controller's upkeep and is cast
// for free with haste when the last counter goes. A card with zero mana value
// is exiled with no counters and so has nothing to count down: it has
// suspend but is never cast by it (the printed result). The second ability
// is the activated "can't be blocked" over a creature with power 3 or less,
// judged on announcement and again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "721e8249-ff1b-4c15-964b-4ba2a96beb75",
		Name:         "The Eleventh Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("The Eleventh Doctor — you may exile a card from your hand with time counters equal to its mana value; it gains suspend",
				eleventhDoctorAsk),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target creature with power 3 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
			Effect:  effects.TargetCantBeBlockedThisTurn("The Eleventh Doctor"),
		}},
	})
}

func eleventhDoctorAsk(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	cards := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		cards = append(cards, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "The Eleventh Doctor — you may exile a card from your hand with time counters equal to its mana value",
		Cards:      cards,
		Min:        0,
		Max:        1,
		ThenKey:    eleventhDoctorExile,
		Carry:      []uuid.UUID{item.Controller},
		LiveZone:   game.ZoneHand,
	})
	return nil
}

// eleventhDoctorExile is the keyed answer: exile the pick, then the counters
// and suspend. An on-disk key: never renamed, never reused.
var eleventhDoctorExile = game.RegisterChooseCardsThen("choose-cards/the-eleventh-doctor-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	for _, id := range r.Picked {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		mv, _ := g.ManaValueForEffect(c)
		if err := g.ExileCardForEffect(id); err != nil {
			return err
		}
		if err := g.GainSuspendForEffect(r.Carry[0], id, mv); err != nil {
			return err
		}
	}
	return nil
})
