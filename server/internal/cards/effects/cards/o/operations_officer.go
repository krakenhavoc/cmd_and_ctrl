package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "c8ed45e6-ae4f-4c44-9f73-45793f30acd6",
		Name:            "Operations Officer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Operations Officer — draw a card", effects.Do(effects.DrawCards{N: 1})), game.Purpose{Draws: 1}),
			effects.WheneverThisAttacks("Operations Officer — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
