package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Elder Dragon War — Enchantment — Saga {2}{R}{R} (#2123):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — This Saga deals 2 damage to each creature and each opponent.
//	 II — Discard any number of cards, then draw that many cards.
//	 III — Create a 4/4 red Dragon creature token with flying."
//
// Read ahead is the engine's (CR 702.155, game/read_ahead.go). Chapter
// I's damage comes from the Saga, to every creature (yours included)
// and to each opponent, not to you. Chapter II draws exactly as many
// cards as were discarded; with an empty hand it does nothing.
//
// No simplification.
func init() {
	// Chapter I is a board wipe for ADR 0126: 2 damage to every
	// creature, the Saga's controller's included.
	wipe := ChapterTrigger(1, SagaChapterLabel("The Elder Dragon War", 1, "2 damage to each creature and each opponent"),
		elderDragonWarDamage)
	wipe.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}}
	Register(Spec{
		OracleID:        "4945f03b-ff99-4923-b62d-3cdb60275ef1",
		Name:            "The Elder Dragon War",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered: []game.TriggeredAbility{
			wipe,
			ChapterTrigger(2, SagaChapterLabel("The Elder Dragon War", 2, "discard any number of cards, then draw that many"),
				elderDragonWarRummage),
			ChapterTrigger(3, SagaChapterLabel("The Elder Dragon War", 3, "create a 4/4 red Dragon with flying"),
				elderDragonWarDragon),
		},
	})
}

// elderDragonWarDamage is chapter I.
func elderDragonWarDamage(g *game.Game, item *game.StackItem) error {
	return damageEachCreatureAndEachOpponent(NewContext(g, item), 2)
}

// elderDragonWarRummage is chapter II: the prompt offers the whole hand,
// any number of it, and the draw is the count actually discarded.
func elderDragonWarRummage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	hand := len(allHandCardIDs(g, item.Controller))
	if hand == 0 {
		return nil
	}
	return b39MayDiscardThenDraw(hand, false,
		"The Elder Dragon War — discard any number of cards, then draw that many",
		func(discarded int) int { return discarded })(ctx)
}

// elderDragonWarDragon is chapter III.
func elderDragonWarDragon(g *game.Game, item *game.StackItem) error {
	return CreateToken{
		Controller: item.Controller,
		Template:   TokenCard("4/4 red Dragon with flying"),
		N:          1,
	}.Apply(NewContext(g, item))
}
