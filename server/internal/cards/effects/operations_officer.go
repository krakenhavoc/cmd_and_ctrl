package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Operations Officer — Creature — Human Officer {3}{W}, 2/2 (Star Trek
// Commander, #4974):
//
//	"Lifelink
//	 When this creature enters, draw a card.
//	 Whenever this creature attacks, put a +1/+1 counter on it."
//
// Lifelink is the printed keyword. The enters trigger draws its
// controller one card; the attack trigger puts one +1/+1 counter on the
// Officer, through the CR 614 window, so counter replacements apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8ed45e6-ae4f-4c44-9f73-45793f30acd6",
		Name:            "Operations Officer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisEnters("Operations Officer — draw a card", Do(DrawCards{N: 1})), game.Purpose{Draws: 1}),
			WheneverThisAttacks("Operations Officer — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
