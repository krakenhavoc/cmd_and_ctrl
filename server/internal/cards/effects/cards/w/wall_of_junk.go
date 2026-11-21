package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Junk — Artifact Creature — Wall {2}, 0/7 (Dominaria Remastered):
//
//	"Defender
//	 When this creature blocks, return it to its owner's hand at end of
//	 combat. (Return it only if it's on the battlefield.)"
//
// A block trigger (once per declaration however many attackers it blocks,
// CR 509.3a) that schedules a delayed trigger at the beginning of this
// combat's end of combat step (CR 511.2, 603.7) naming this very object
// (the epoch, CR 400.7). It returns the Wall only if that object is still
// on the battlefield then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22b0de2d-47a8-4269-ba21-4bc2b4b10b40",
		Name:            "Wall of Junk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Wall of Junk — return it to its owner's hand at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ScheduleDelayedTrigger{
						At:     game.StepEndCombat,
						Label:  "Wall of Junk — return it to its owner's hand",
						Body:   effects.S82GapB134ReturnSourceToHandBody,
						Params: game.EffectParams{Object: game.ObjectRef{ID: item.SourceCardID, Epoch: g.ObjectEpochForEffect(item.SourceCardID)}},
					}.Apply(ctx)
				}),
		},
	})
}
