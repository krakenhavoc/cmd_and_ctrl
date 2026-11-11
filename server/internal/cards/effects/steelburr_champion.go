package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steelburr Champion — Creature — Mouse Soldier {2}{W}, 1/1 (Bloomburrow
// Commander, #5368):
//
//	"Offspring {1}{W} ... Vigilance
//	 Whenever an opponent casts a noncreature spell, put a +1/+1 counter
//	 on this creature."
//
// Offspring is the shared keyword; the token copy has the trigger too.
func init() {
	Register(Spec{
		OracleID:        "d1f3d5ca-cb17-400f-8ed9-3b9a4f151b62",
		Name:            "Steelburr Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		OptionalCosts:   []game.AdditionalCost{Offspring("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Steelburr Champion"),
			On(game.EventCast, cmdr2S18OpponentCastsNoncreature,
				"Steelburr Champion — a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
