package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghost-Spider — the back face of Gwen Stacy // Ghost-Spider (#3987).
//
//	Ghost-Spider                    {2}{U}{R}{W}   4/4
//	Legendary Creature — Spider Human Hero
//	  Flying, vigilance, haste
//	  Whenever you play a land from exile or cast a spell from exile,
//	  put a +1/+1 counter on Ghost-Spider.
//	  Remove two counters from Ghost-Spider: Exile the top card of your
//	  library. You may play that card this turn.
//
// Reached by Gwen Stacy's transform (gwen_stacy.go) or by casting this
// face of the modal double-faced card.
//
// The trigger is Urianger Augurelt's and Prosper's read of the two
// events: a spell cast from exile (EventCast with an exile origin) or a
// land played from exile. Any permission counts, including the cards
// this face's own ability exiles.
//
// "Remove two counters from Ghost-Spider" prints no kind, so it is two
// counters of whatever kinds she holds: two +1/+1, or a +1/+1 and a
// shield counter, as the activator chooses while paying (CR 602.2b,
// 118.3). RemoveCountersOfAnyKindsFromThis asks the kind per part, and
// the client offers one row per kind she holds.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        gwenStacyOracle + "#1",
		Name:            "Ghost-Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20PlayedACardFromExile(ev, source, g)
			}, "Ghost-Spider — put a +1/+1 counter on Ghost-Spider", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Remove two counters from Ghost-Spider: Exile the top card of your library. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    RemoveCountersOfAnyKindsFromThis(2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
