package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Royal Trooper — Creature — Human Soldier {2}{W}, 2/2 (Starter 1999,
// #5620):
//
//	"Whenever this creature blocks, it gets +2/+2 until end of turn."
//
// Giant Badger's text: "whenever this creature blocks" names no attacker,
// so it triggers once however many attackers the Trooper blocks (CR
// 509.3a, selfBlocksOnce). The pump is the source's own and ends in the
// cleanup step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e31510d5-b74e-4a8d-94e8-8bd098cac3c7",
		Name:         "Royal Trooper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Royal Trooper — gets +2/+2 until end of turn",
				thisGetsUntilEndOfTurn(2, 2, "Royal Trooper — +2/+2")),
		},
	})
}
