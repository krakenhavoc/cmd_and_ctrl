package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Descendants' Fury — Enchantment {3}{R}:
//
//	"Whenever one or more creatures you control deal combat damage to a
//	 player, you may sacrifice one of them. If you do, reveal cards from the
//	 top of your library until you reveal a creature card that shares a
//	 creature type with the sacrificed creature. Put that card onto the
//	 battlefield and the rest on the bottom of your library in a random order."
//
// "One of them" is a creature of the batch: the controller's creatures that
// dealt combat damage to that player in the same damage step (Event.Batch).
// The sacrificed creature's creature types ride the tail as a word, read
// before it goes (CR 608.2h).
func init() {
	Register(Spec{
		OracleID:     "97079964-7a2e-42c5-ad4a-9015c18a1e97",
		Name:         "Descendants' Fury",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Descendants' Fury — you may sacrifice one of them; if you do, reveal until a creature card that shares a creature type and put it onto the battlefield",
				descendantsFuryOffer),
		},
	})
}

// descendantsFuryOffer asks which of the creatures that connected to
// sacrifice (nothing is a legal answer).
func descendantsFuryOffer(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	var candidates []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, prior := range g.Events {
		if prior.Batch != ev.Batch || prior.Kind != game.EventDealDamage || !prior.Combat ||
			prior.Amount <= 0 || prior.Target != ev.Target || seen[prior.Source] {
			continue
		}
		seen[prior.Source] = true
		c, ok := g.LookupCardForEffect(prior.Source)
		if ok && c.IsCreature() && c.Controller == item.Controller && g.Battlefield.Contains(prior.Source) &&
			!g.CantBeSacrificedForEffect(prior.Source) {
			candidates = append(candidates, prior.Source)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     ctx.Source(),
		Question:   "Descendants' Fury — you may sacrifice one of the creatures that dealt combat damage",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    descendantsFurySacrificePicked,
	})
	return nil
}

// descendantsFurySacrificePicked sacrifices the pick, carrying its creature
// types to the reveal. An on-disk key: never renamed, never reused.
var descendantsFurySacrificePicked = game.RegisterChooseCardsThen("choose-cards/descendants-fury-sacrifice",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		c, ok := g.LookupCardForEffect(r.Picked[0])
		if !ok {
			return nil
		}
		word := cmdr2B03TypesWord(c.Effective().Subtypes, game.HasAllCreatureTypes(&c))
		return g.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked, game.KeyedTail{
			Ref: descendantsFurySacrificed, Player: r.Chooser, Source: r.Source, Word: word,
		})
	})

// descendantsFurySacrificed is "if you do": the creature really went, so
// reveal until a creature card shares a type with it. An on-disk key.
var descendantsFurySacrificed = CardTail("then/descendants-fury-sacrificed", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.Word == "" {
		return nil
	}
	return RevealUntilThenPutOntoBattlefield{
		Match:  cmdr2B03SharesTypeWith(o.Tail.Word),
		Reason: "Descendants' Fury — reveal until a creature card that shares a creature type",
	}.Apply(ctx)
})
