package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Love Song of Night and Day — Enchantment — Saga {2}{W} (#2123):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — You and target opponent each draw two cards.
//	 II — Create a 1/1 white Bird creature token with flying.
//	 III — Put a +1/+1 counter on each of up to two target creatures."
//
// Read ahead is the engine's (CR 702.155, game/read_ahead.go). Chapter
// I draws for the Saga's controller first and then the opponent, the
// printed order; an opponent no longer a legal target draws nothing.
// Chapter III may target none, one or two creatures, and counts only
// those still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b87b4cd-4fb9-4bb3-b4c0-de1fcc6f44a2",
		Name:            "Love Song of Night and Day",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("Love Song of Night and Day", 1, "you and target opponent each draw two cards"),
				TargetPlayer("target opponent", Opponent()),
				loveSongEachDrawTwo),
			ChapterTrigger(2, SagaChapterLabel("Love Song of Night and Day", 2, "create a 1/1 white Bird with flying"),
				loveSongBird),
			ChapterTriggerTargeting(3, SagaChapterLabel("Love Song of Night and Day", 3, "a +1/+1 counter on each of up to two creatures"),
				TargetCreature("up to two target creatures").WithCount(0, 2),
				plusOneCounterOnChosenTargets),
		},
	})
}

// loveSongEachDrawTwo is chapter I: two for the controller, then two
// for the targeted opponent if it is still a legal target.
func loveSongEachDrawTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return DrawCards{Player: t.ID, N: 2}.Apply(ctx)
		}
	}
	return nil
}

// loveSongBird is chapter II.
func loveSongBird(g *game.Game, item *game.StackItem) error {
	return CreateToken{
		Controller: item.Controller,
		Template:   TokenCard("1/1 white Bird with flying"),
		N:          1,
	}.Apply(NewContext(g, item))
}
