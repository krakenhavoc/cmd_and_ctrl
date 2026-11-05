package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang's Journey — Sorcery — Lesson {2} (Avatar: The Last Airbender,
// #4459):
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 Search your library for a basic land card. If this spell was kicked,
//	 instead search your library for a basic land card and a Shrine card.
//	 Reveal those cards, put them into your hand, then shuffle.
//	 You gain 2 life."
//
// One search either way. Unkicked it finds one basic land card; kicked it
// finds up to two cards in one prompt, held by a registered set rule
// (aangsJourneyBasicAndShrine) to one basic land card and one Shrine card
// (Corpse Harvester's "a Zombie card and a Swamp card"), so the searcher
// may find either half alone, or nothing (CR 701.23b: a search for a
// stated quality may fail to find). The life comes after the shuffle, as
// printed, from the search's continuation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ddf461b9-a205-4dc7-a9c3-c047b1ff709f",
		Name:          "Aang's Journey",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tutors: 1},
		OptionalCosts: []game.AdditionalCost{Kicker("{2}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			search := SearchLibrary{
				Player:    item.Controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Aang's Journey — a basic land card",
				Source:    item.SourceCardID,
				ThenTail:  ctx.TailN(thenChangeLife, 2),
			}
			if ctx.WasKicked() {
				search.Predicate = func(c game.Card) bool { return IsBasicLand(c) || aangsJourneyShrine(c) }
				search.Limit = 2
				search.Rule = aangsJourneyBasicAndShrine
				search.Reason = "Aang's Journey — a basic land card and a Shrine card"
			}
			return search.Apply(ctx)
		},
	})
}

// aangsJourneyShrine is "a Shrine card": the enchantment subtype.
func aangsJourneyShrine(c game.Card) bool { return c.HasSubtypeWord("Shrine") }

// aangsJourneyBasicAndShrine is the kicked search's set rule: at most one
// basic land card and at most one Shrine card, each pick filling a
// different half (a card that is both may fill either). An on-disk key.
var aangsJourneyBasicAndShrine = game.RegisterSearchRule("search/aangs-journey-basic-land-and-shrine", func(picked []game.Card, _ int) bool {
	switch len(picked) {
	case 0:
		return true
	case 1:
		return IsBasicLand(picked[0]) || aangsJourneyShrine(picked[0])
	case 2:
		a, b := picked[0], picked[1]
		return IsBasicLand(a) && aangsJourneyShrine(b) || aangsJourneyShrine(a) && IsBasicLand(b)
	}
	return false
})
