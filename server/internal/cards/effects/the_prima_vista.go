package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Prima Vista — Legendary Artifact — Vehicle {4}{U}, 5/3 (Final
// Fantasy, #4462):
//
//	"Flying
//	 Whenever you cast a noncreature spell, if at least four mana was
//	 spent to cast it, The Prima Vista becomes an artifact creature
//	 until end of turn.
//	 Crew 2"
//
// Blazing Bomb's cast trigger over the spell's payment record, with the
// crew ability's own resolution as its body: the same "becomes an
// artifact creature until end of turn" (CR 702.122), pinned to this
// object so a Vehicle that left and came back is not animated
// (CR 400.7). Becoming a creature this way is not crewing, so nothing
// that watches for a Vehicle becoming crewed sees it.
//
// No simplification.
func init() {
	animate := On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return YouCast(Noncreature())(ev, source, lki, g) && g.StackItemPaidForEffect(ev.CardID).Spent().Total() >= 4
	}, "The Prima Vista — becomes an artifact creature until end of turn",
		Do(BecomeCreatureUntilEOT{Label: "The Prima Vista — becomes an artifact creature"}))
	Register(Spec{
		OracleID:        "a09d9754-a321-4018-a237-e613e915b642",
		Name:            "The Prima Vista",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{animate},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("The Prima Vista"),
		}},
	})
}
