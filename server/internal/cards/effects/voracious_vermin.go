package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Vermin — Creature — Rat {2}{B}, 2/1 (Wilds of Eldraine,
// #4471):
//
//	"When this creature enters, create a 1/1 black Rat creature token
//	 with "This token can't block."
//	 Whenever another creature you control dies, put a +1/+1 counter on
//	 this creature."
//
// The entry makes the catalog's can't-block Rat (Pest Problem's). The
// second trigger is Feral Ghoul's: another creature that died under your
// control, read as it last existed (CR 603.10a), so the Rat token dying
// counts. The counter goes on the Vermin only while it is still the same
// permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d36d80e-9934-4b56-bf3c-2b18a6da3d2d",
		Name:         "Voracious Vermin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Voracious Vermin — create a 1/1 black Rat creature token with \"This token can't block.\"",
				Do(CreateToken{Template: CantBlockRatToken(), N: 1})),
			On(game.EventLTB, AllOf(AnotherCreatureDied, ACreatureYouControlDied),
				"Voracious Vermin — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
