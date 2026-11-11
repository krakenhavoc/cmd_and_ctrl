package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tishana, Voice of Thunder — Legendary Creature — Merfolk Shaman
// {5}{G}{U}, */* (Lost Caverns of Ixalan Commander, #5368):
//
//	"Tishana's power and toughness are each equal to the number of cards
//	 in your hand.
//	 You have no maximum hand size.
//	 When Tishana enters, draw a card for each creature you control."
//
// The P/T is a characteristic-defining ability in layer 7a (CR 604.3),
// the two halves Psychosis Crawler and Body of Knowledge share. The draw
// is counted as the trigger resolves, Tishana herself included while she
// is on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "479dd04b-0048-446a-9202-a4dc9472e5c3",
		Name:          "Tishana, Voice of Thunder",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
		Purpose:       game.Purpose{Draws: 1},
		Static: []game.StaticAbility{
			powerIsCardsInYourHand(),
			cmdr2B08ToughnessIsCardsInYourHand(),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tishana — draw a card for each creature you control",
				b34DrawPerCreatureYouControlMatching(func(game.Card) bool { return true })),
		},
	})
}
