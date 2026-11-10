package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Body of Knowledge — Creature — Avatar {3}{U}{U}, */* (Lost Caverns of Ixalan
// Commander):
//
//	"Body of Knowledge's power and toughness are each equal to the number of
//	 cards in your hand.
//	 You have no maximum hand size.
//	 Whenever this creature is dealt damage, draw that many cards."
//
// A characteristic-defining ability in layer 7a (CR 604.3) read through two
// statics, one per side. A creature that is 0 toughness dies as a state-based
// action (CR 704.5f) like any other.
func init() {
	Register(Spec{
		OracleID:      "7fc6a5c1-f0da-42f6-93d0-6d94467e2ac0",
		Name:          "Body of Knowledge",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
		Static: []game.StaticAbility{
			powerIsCardsInYourHand(),
			cmdr2B08ToughnessIsCardsInYourHand(),
		},
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Body of Knowledge — draw that many cards", cmdr2B08DrawThatMany),
		},
	})
}
