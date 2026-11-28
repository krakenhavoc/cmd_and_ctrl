package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Varchild's War-Riders — Creature — Human Warrior {1}{R} 3/4 (Alliances,
// S82 OLD batch 818):
//
//	"Cumulative upkeep—Have an opponent create a 1/1 red Survivor creature
//	 token.
//	 Trample; rampage 1"
//
// The payment is the one #6189 built: for each age counter the controller
// names an opponent, who creates the Survivor under their own control (CR
// 111.2); one opponent may be named more than once. The token row is filed
// under the payment's key at init.
func init() {
	game.RegisterPayActionToken(oldB818SurvivorKey, effects.TokenCard("1/1 red Survivor"))
	effects.Register(effects.Spec{
		OracleID:        "258921f2-df26-4d8a-ad34-8f44d0158ae7",
		Name:            "Varchild's War-Riders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "rampage 1"},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Varchild's War-Riders — cumulative upkeep: have an opponent create a 1/1 red Survivor creature token",
				effects.OpponentCreatesTokenPayment(1, oldB818SurvivorKey, "1/1 red Survivor creature token"), "", ""),
		},
	})
}

const oldB818SurvivorKey = "varchilds-war-riders/survivor"
