package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earth King's Lieutenant — Creature — Human Soldier Ally {G}{W}, 1/1
// (Avatar: The Last Airbender, #4459):
//
//	"Trample
//	 When this creature enters, put a +1/+1 counter on each other Ally
//	 creature you control.
//	 Whenever another Ally you control enters, put a +1/+1 counter on
//	 this creature."
//
// The enters trigger counts the other Ally creatures as it resolves
// (Nazgûl's putAPlusOneCounterOnEachYouControl); the second is Avatar
// Enthusiasts' trigger. "Ally" is a word of the card's text (CR 612.2),
// read post-layer, so a changeling counts. An Ally entering at the same
// time as the Lieutenant triggers both and gets a counter too (the
// ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9da9248d-1201-447f-b6c2-2b64af4f71c4",
		Name:            "Earth King's Lieutenant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Earth King's Lieutenant — put a +1/+1 counter on each other Ally creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ally := NewContext(g, item).TextSubtype("Ally")
					return putAPlusOneCounterOnEachYouControl(And(Creature(), Subtype(ally), OtherThan(item.SourceCardID)))(g, item)
				}),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, "Earth King's Lieutenant — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
