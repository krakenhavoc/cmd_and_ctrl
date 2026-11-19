package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overwhelming Forces — Sorcery {6}{B}{B} (Breaking News, #5929):
//
//	"Destroy all creatures target opponent controls. Draw a card for each
//	 creature destroyed this way."
//
// One target, an opponent. The set is every creature that player controls
// as the spell resolves, and the draw counts the ones actually destroyed: an
// indestructible creature, or one saved by a regeneration shield, is not
// counted (CR 701.8, CR 702.12b). The draw is the sweep's keyed tail, so a
// table waiting on a prompt the wipe asks (a commander's zone choice) is a
// restore point. An opponent who has left the game before it resolves makes
// the target illegal and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdccceb6-ad3c-4a75-a8a6-ca796ede4185",
		Name:         "Overwhelming Forces",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true},
			Answers: game.AnswerRemove,
		},
		Targets: TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opp := TargetedPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			return DestroyAllMatching{
				Match:    And(Creature(), ControlledBy(opp)),
				ThenTail: ctx.Tail(overwhelmingForcesDraw),
			}.Apply(ctx)
		},
	})
}

// overwhelmingForcesDraw is "draw a card for each creature destroyed this
// way". An on-disk key: never renamed, never reused.
var overwhelmingForcesDraw = MassTail("mass/overwhelming-forces-draw", func(ctx *Context, _ []game.Card, destroyed int) error {
	return DrawCards{N: destroyed}.Apply(ctx)
})
