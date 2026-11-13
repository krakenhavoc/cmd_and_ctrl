package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Badger — Creature — Badger {1}{G}{G}, 2/2 (Eighth Edition,
// #5607):
//
//	"Whenever this creature blocks, it gets +2/+2 until end of turn."
//
// Burning Sun Cavalry's block half with no condition: "whenever this
// creature blocks" names no attacker, so it triggers once however many
// attackers the Badger blocks (CR 509.3a, selfBlocksOnce). The pump is
// the source's own and ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4109138-5d69-4510-a228-7c1f88c0c8c0",
		Name:         "Giant Badger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return selfBlocksOnce(ev, source)
			}, "Giant Badger — gets +2/+2 until end of turn",
				thisGetsUntilEndOfTurn(2, 2, "Giant Badger — +2/+2")),
		},
	})
}
