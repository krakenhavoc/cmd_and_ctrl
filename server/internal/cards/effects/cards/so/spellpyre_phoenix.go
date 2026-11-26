package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spellpyre Phoenix — {3}{R}{R} Creature — Phoenix 4/2 (Commander 2020):
//
//	"Flying
//	 When this creature enters, you may return target instant or sorcery card
//	 with a cycling ability from your graveyard to your hand.
//	 At the beginning of each end step, if you cycled two or more cards this
//	 turn, return this card from your graveyard to your hand."
//
// The entry trigger is targeted, and returns the card if it is still a legal
// target as it resolves (CR 608.2b); a "may" whose only outcome is a card in
// hand is not asked. The end-step trigger functions from the graveyard
// (Zones, CR 113.6k) and is an intervening "if" (CR 603.4): the number of
// cards cycled is the controller's EventCycle events this turn (CR 702.29b,
// typecycling included), checked as it would trigger and again as it
// resolves. It fires at every player's end step.
func init() {
	enters := effects.WhenThisEnters("Spellpyre Phoenix — return target instant or sorcery card with a cycling ability from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
		})
	enters.Targets = effects.TargetCardInGraveyard("target instant or sorcery card with a cycling ability from your graveyard",
		effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn(), effects.GapB158HasCyclingAbility)
	back := effects.InterveningIfYou("you cycled two or more cards this turn", spellpyrePhoenixCycledTwo,
		effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return source != nil && ev.Kind == game.EventBeginEndStep
		}, "Spellpyre Phoenix — return this card from your graveyard to your hand", effects.ReturnThisCardFromYourGraveyardToYourHand))
	back.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:        "a86710da-e498-4ca1-832c-8ba4abd6fa46",
		Name:            "Spellpyre Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters, back},
	})
}

// spellpyrePhoenixCycledTwo reads "you cycled two or more cards this turn".
func spellpyrePhoenixCycledTwo(g *game.Game, you uuid.UUID) bool {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventCycle && ev.Actor == you {
			n++
		}
	}
	return n >= 2
}
