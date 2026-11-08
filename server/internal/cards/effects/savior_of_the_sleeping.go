package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savior of the Sleeping — Creature — Human Knight {2}{W}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"Vigilance
//	 Whenever an enchantment you control is put into a graveyard from the
//	 battlefield, put a +1/+1 counter on this creature."
//
// Vigilance is the engine's keyword. The trigger is Knight of Doves'
// (socB13EnchantmentYouControlWasPutIntoAGraveyard): the enchantment is
// read as it last existed on the battlefield and under whose control (CR
// 603.10a), so an Aura that falls off its creature and a Role token both
// count, and an enchantment an opponent controlled does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "412eb5ec-bf13-45fb-94b3-c41e62b19d1d",
		Name:            "Savior of the Sleeping",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, socB13EnchantmentYouControlWasPutIntoAGraveyard,
				"Savior of the Sleeping — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
