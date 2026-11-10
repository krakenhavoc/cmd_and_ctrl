package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rediscover the Way — Enchantment — Saga {U}{R}{W} (Tarkir: Dragonstorm,
// #4463):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Look at the top three cards of your library. Put one of them
//	 into your hand and the rest on the bottom of your library in any
//	 order.
//	 III — Whenever you cast a noncreature spell this turn, target creature
//	 you control gains double strike until end of turn."
//
// Chapters I and II are Impulse's dig over three cards. Chapter III is
// Origin of Thor's repeating delayed trigger (CR 603.7b) over Good King
// Mog's "you cast a noncreature spell" condition, with its own target
// clause announced as each instance goes on the stack. It outlives the
// Saga (the ruling): it triggers for every noncreature spell you cast for
// the rest of the turn, and each gives double strike until end of turn to
// the creature chosen then, if it is still a legal target.
//
// No simplification.
func init() {
	const name = "Rediscover the Way"
	dig := LookAtTopThenTakeOneRestOnBottom(3, "Rediscover the Way — put one of them into your hand")
	Register(Spec{
		OracleID:     "680a8e14-053f-4cde-a060-e1b36873cf14",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "look at the top three cards of your library; put one into your hand and the rest on the bottom in any order"), dig),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "look at the top three cards of your library; put one into your hand and the rest on the bottom in any order"), dig),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "whenever you cast a noncreature spell this turn, target creature you control gains double strike until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     "Rediscover the Way — target creature you control gains double strike until end of turn",
						On:        []game.EventKind{game.EventCast},
						Condition: mogNoncreatureCastCondition,
						Body:      rediscoverTheWayBody,
						Repeats:   true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// rediscoverTheWayBody is chapter III's delayed trigger; its target clause
// is announced as each instance goes on the stack. An on-disk key: never
// renamed, never reused.
var rediscoverTheWayBody = game.TargetedDelayedBody("rediscover-the-way/double-strike",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"},
			Label: "Rediscover the Way — double strike"}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))
