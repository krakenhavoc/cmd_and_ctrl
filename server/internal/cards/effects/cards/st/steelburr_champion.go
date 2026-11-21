package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steelburr Champion — Creature — Mouse Soldier {2}{W}, 1/1 (Bloomburrow
// Commander, #5368):
//
//	"Offspring {1}{W} ... Vigilance
//	 Whenever an opponent casts a noncreature spell, put a +1/+1 counter
//	 on this creature."
//
// Offspring is the shared keyword; the token copy has the trigger too.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d1f3d5ca-cb17-400f-8ed9-3b9a4f151b62",
		Name:            "Steelburr Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		OptionalCosts:   []game.AdditionalCost{effects.Offspring("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.OffspringToken("Steelburr Champion"),
			effects.On(game.EventCast, effects.Cmdr2S18OpponentCastsNoncreature,
				"Steelburr Champion — a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
