package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Sleuth — Creature — Angel Advisor {2}{W}, 2/3:
//
//	"Flying
//	 Whenever another permanent you control leaves the battlefield, if
//	 it had counters on it, investigate."
//
// The counters are the permanent's as it last existed (CR 603.10a),
// the record the engine writes as it leaves; the controller is the one
// it had as it left.
func init() {
	Register(Spec{
		OracleID:        "2579e9e0-b00f-4a1b-8fce-15d06bb7fe5a",
		Name:            "Angelic Sleuth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b14PermanentYouControlLeft(ev, source, g)
			}, "Angelic Sleuth — investigate", Do(cmdrB1Investigate())),
		},
	})
}
