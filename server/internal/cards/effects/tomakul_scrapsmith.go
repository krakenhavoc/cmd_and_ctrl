package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomakul Scrapsmith — Creature — Human Artificer {2}{R}, 2/1 (The Brothers' War):
//
//	"When this creature enters, mill three cards. You may put an artifact card from
//	 among the cards milled this way into your hand. If you don't, put a
//	 +1/+1 counter on this creature."
//
// Ainok Wayfarer's mill-then-take shape (bro6MillThreeTakeOne), offering
// an artifact card from the cards this mill put into the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1559911-6621-4fd6-8b3e-16e82baf849e",
		Name:         "Tomakul Scrapsmith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			bro6MillThreeTakeOne("Tomakul Scrapsmith", "tomakul-scrapsmith", "an artifact card", Artifact()),
		},
	})
}
