package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Cruelty of Gix — Enchantment — Saga {3}{B}{B} (#2123):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — Target opponent reveals their hand. You choose a creature or
//	     planeswalker card from it. That player discards that card.
//	 II — Search your library for a card, put that card into your hand,
//	      then shuffle. You lose 3 life.
//	 III — Put target creature card from a graveyard onto the
//	       battlefield under your control."
//
// Read ahead is the engine's (CR 702.155, game/read_ahead.go): the
// printed keyword asks for the starting chapter as the Saga enters, and
// the chapters it skips never trigger. Chapter I is Thoughtseize's pick
// (ADR 0116), chosen by the Saga's controller (CR 113.8). Chapter II is
// Grim Tutor's search and life loss. Chapter III says "a graveyard",
// so any player's creature card may be chosen, and it enters under the
// Saga's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2b5c62af-af5f-4beb-9688-19604ca46918",
		Name:            "The Cruelty of Gix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("The Cruelty of Gix", 1, "target opponent reveals their hand"),
				TargetPlayer("target opponent", Opponent()),
				TargetRevealsYouChooseDiscardAbility(Or(Creature(), Planeswalker()), "creature or planeswalker card")),
			ChapterTrigger(2, SagaChapterLabel("The Cruelty of Gix", 2, "search your library for a card, then lose 3 life"),
				crueltyOfGixTutor),
			ChapterTriggerTargeting(3, SagaChapterLabel("The Cruelty of Gix", 3, "reanimate a creature under your control"),
				TargetCardInGraveyard("target creature card in a graveyard", Creature()),
				eldestRebornReanimate),
		},
	})
}

// crueltyOfGixTutor is chapter II.
func crueltyOfGixTutor(g *game.Game, item *game.StackItem) error {
	return searchForACardThenLoseLife(item, NewContext(g, item),
		"The Cruelty of Gix — search your library for a card", 3)
}
