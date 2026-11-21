package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver Surfer, Cosmic Voyager — {4}{U}{U} Legendary Creature — Alien Hero 5/5
// (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 Flying
//	 When Silver Surfer enters, exile any number of other target permanents you
//	 control. Return those cards to the battlefield under their owner's
//	 control at the beginning of the next end step. If a land enters this way,
//	 it enters tapped."
//
// Disruptor of Currents's "other target" clause with a count of any number,
// then the shared blink (exile all at once, one delayed trigger that returns
// them together as one simultaneous entry, CR 603.6a). The land clause is the
// return's own "enters tapped" (CR 614.1c), given card by card by what each
// would be on the battlefield (CR 614.12): ZoneEntryOptions.EntersTappedIfType,
// the tapped sibling of the counters clause Semester's End uses. Splitting the
// lands from the rest into two returns would be two events, which the card
// does not say. A token exiled this way ceases to exist and does not return
// (CR 111.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ce1f550-d044-4cb1-991b-e98449f74507",
		Name:            "Silver Surfer, Cosmic Voyager",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.WhenThisEnters("Silver Surfer, Cosmic Voyager — exile any number of other target permanents you control; return them at the beginning of the next end step, lands tapped",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ExileTargetsThenScheduleReturnBody(effects.NewContext(g, item),
							"Silver Surfer, Cosmic Voyager — return the exiled permanents; lands enter tapped",
							effects.ReturnExiledLandsTappedBody)
					}),
				effects.AnotherTarget(func(other effects.CardPredicate) *game.TargetSpec {
					return effects.TargetPermanent("any number of other target permanents you control", other, effects.YouControl()).WithCount(0, 0)
				})),
		},
	})
}
