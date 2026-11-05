package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Naked Time — Enchantment — Saga {1}{R} (Star Trek, #4973):
//
//	"I — Search your library for a basic Mountain card, put it onto the
//	     battlefield tapped, then shuffle.
//	 II — Whenever a creature you control enters this turn, put a +1/+1
//	      counter on it and it gains haste until end of turn.
//	 III — Sacrifice a land."
//
// Chapter I is the basic-Mountain fetch (IsBasicLandWithSubtype): the
// library is shuffled even when none is found. Chapter II creates a
// delayed triggered ability for the rest of the turn (CR 603.7b, First
// Day of Class's shape): it triggers for each creature that enters under
// your control after the chapter resolves, and the counter and the haste
// go to that creature only if it is still the same object on the
// battlefield (CR 603.7c). Chapter III asks you to sacrifice a land of
// your choice; with none there is nothing to sacrifice (CR 701.21a). The
// Saga is sacrificed after III by the lore-counter rule (CR 714.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2cf79dd0-c9db-403e-969d-6c09e7aaa575",
		Name:         "The Naked Time",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Naked Time — I: search for a basic Mountain card, put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLandWithSubtype("Mountain"),
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "The Naked Time — a basic Mountain card, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(2, "The Naked Time — II: creatures you control that enter this turn get a +1/+1 counter and haste",
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     "The Naked Time — put a +1/+1 counter on it and it gains haste until end of turn",
						On:        []game.EventKind{game.EventETB},
						Condition: eachCreatureYouControlEntersCondition,
						Body:      counterAndHasteForThePayloadBody,
						Repeats:   true,
					}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(3, "The Naked Time — III: sacrifice a land",
				func(g *game.Game, item *game.StackItem) error {
					return SacrificeChoice{
						Player:     item.Controller,
						Candidates: permanentsControlledByMatching(g, item.Controller, Land()),
						Question:   "The Naked Time — sacrifice a land",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
