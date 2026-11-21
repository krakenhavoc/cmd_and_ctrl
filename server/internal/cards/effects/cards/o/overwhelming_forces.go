package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "bdccceb6-ad3c-4a75-a8a6-ca796ede4185",
		Name:         "Overwhelming Forces",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true},
			Answers: game.AnswerRemove,
		},
		Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			opp := effects.TargetedPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			return effects.DestroyAllMatching{
				Match:    effects.And(effects.Creature(), effects.ControlledBy(opp)),
				ThenTail: ctx.Tail(overwhelmingForcesDraw),
			}.Apply(ctx)
		},
	})
}

// overwhelmingForcesDraw is "draw a card for each creature destroyed this
// way". An on-disk key: never renamed, never reused.
var overwhelmingForcesDraw = effects.MassTail("mass/overwhelming-forces-draw", func(ctx *effects.Context, _ []game.Card, destroyed int) error {
	return effects.DrawCards{N: destroyed}.Apply(ctx)
})
