package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian Kirin — Creature — Kirin {3}{W}, 2/3 (Modern Horizons 2,
// #5501):
//
//	"Flying
//	 Whenever another creature you control dies, put a +1/+1 counter on
//	 this creature."
//
// Feral Ghoul's counter trigger on a flyer: one trigger per other
// creature of yours that dies, tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e9dbfda3-e162-49f0-b9bf-f8213e27c17d",
		Name:            "Guardian Kirin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AllOf(AnotherCreatureDied, ACreatureYouControlDied),
				"Guardian Kirin — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
