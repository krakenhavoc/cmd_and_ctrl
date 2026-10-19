package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yorion, Sky Nomad — Legendary Creature — Bird Serpent {3}{W/U}{W/U},
// 4/5:
//
//	"Companion — Your starting deck contains at least twenty cards more
//	 than the minimum deck size. (If this card is your chosen companion,
//	 you may put it into your hand from outside the game for {3} as a
//	 sorcery.)
//	 Flying
//	 When Yorion enters, exile any number of other nonland permanents you
//	 own and control. Return those cards to the battlefield at the
//	 beginning of the next end step."
//
// The permanents are CHOSEN as the trigger resolves, not targeted, so
// hexproof and shroud do not matter and nothing can respond to the
// choice. Any number, none included; Yorion itself is never offered,
// and neither is a permanent you control but don't own, or one you own
// that someone else controls. The pick is a keyed choose_cards prompt
// (yorionChosen), so a table waiting on the answer is still a restore
// point. The continuation re-checks each pick (it may have left, or
// changed control, between the question and the answer), then exiles
// every one still eligible as ONE event and schedules the CR 603.7
// delayed trigger over the cards that reached exile
// (exileCardsThenScheduleReturn). They come back together under their
// owner's control, which is you. A token is gone for good
// (CR 111.8).
//
// Companion (ADR 0172) is not this file's business: the condition is
// checked by internal/deck's companion registry. In Commander the
// minimum deck size is also the maximum (CR 903.5a, exactly 100), so no
// starting deck is twenty cards over it and the deck check always
// refuses Yorion as a companion, saying why. In the 99 Yorion is an
// ordinary card, and its companion ability does nothing, as printed.
//
// No simplification. Roon deck request (#4128).
func init() {
	Register(Spec{
		OracleID:        "e15504a7-2b67-4185-b916-172145f10b19",
		Name:            "Yorion, Sky Nomad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Yorion, Sky Nomad — exile your other nonland permanents until the next end step", yorionEnters),
		},
	})
}

// yorionEligible is "other nonland permanent you own and control",
// asked of a battlefield card for the Yorion object `self` controlled
// by `you`.
func yorionEligible(c game.Card, you, self uuid.UUID) bool {
	return c.InstanceID != self && c.Controller == you && c.Owner == you && !c.IsLand()
}

func yorionEnters(g *game.Game, item *game.StackItem) error {
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if yorionEligible(c, item.Controller, item.SourceCardID) {
			mine = append(mine, c.InstanceID)
		}
	}
	if len(mine) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Yorion, Sky Nomad — exile any number of other nonland permanents you own and control; they return at the beginning of the next end step",
		Cards:    mine,
		Min:      0,
		Max:      len(mine), // "any number"
		Zone:     game.ZoneBattlefield,
		ThenKey:  yorionChosen,
	})
	return nil
}

// yorionChosen is the keyed answer to "which permanents". Its key is an
// on-disk identity: never renamed, never reused.
var yorionChosen = game.RegisterChooseCardsThen("choose-cards/yorion-sky-nomad-exile", yorionExileChosen)

// yorionExileChosen exiles every pick that is still an other nonland
// permanent the chooser owns and controls, and schedules their return.
func yorionExileChosen(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	var ids []uuid.UUID
	for _, id := range r.Picked {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !g.Battlefield.Contains(id) || !yorionEligible(c, r.Chooser, r.Source) {
			continue
		}
		ids = append(ids, id)
	}
	ctx := NewContext(g, &game.StackItem{
		Kind:         game.StackItemTriggered,
		Controller:   r.Chooser,
		Owner:        r.Chooser,
		SourceCardID: r.Source,
	})
	return exileCardsThenScheduleReturn(ctx, ids, "Yorion, Sky Nomad — return the exiled permanents")
}
