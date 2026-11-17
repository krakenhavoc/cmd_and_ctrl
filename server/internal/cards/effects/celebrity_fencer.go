package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celebrity Fencer — Creature — Elf Druid {3}{W}, 3/2 (Streets of New
// Capenna, #5491):
//
//	"Alliance — Whenever another creature you control enters, put a +1/+1
//	 counter on this creature."
//
// Alliance is "another creature you control enters" (Boss's Chauffeur's
// trigger): a token counts, an opponent's creature and the Fencer's own
// entry do not. The counter goes on the Fencer only if it is still the same
// permanent when the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39f7e59e-28dd-477e-be38-3fa515838ca4",
		Name:         "Celebrity Fencer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Celebrity Fencer — alliance: put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
