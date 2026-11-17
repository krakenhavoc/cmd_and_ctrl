package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallaji Archaeologist — Creature — Human Scout {1}{U}, 0/3 (The Brothers' War):
//
//	"When this creature enters, mill three cards. You may put a noncreature, nonland card from
//	 among the cards milled this way into your hand. If you don't, put a
//	 +1/+1 counter on this creature."
//
// Ainok Wayfarer's mill-then-take shape (bro6MillThreeTakeOne), offering
// a noncreature, nonland card from the cards this mill put into the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf037457-82e3-45fd-bc17-5bc25e66c15e",
		Name:         "Fallaji Archaeologist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			bro6MillThreeTakeOne("Fallaji Archaeologist", "fallaji-archaeologist", "a noncreature, nonland card", And(Not(Creature()), Not(Land()))),
		},
	})
}
