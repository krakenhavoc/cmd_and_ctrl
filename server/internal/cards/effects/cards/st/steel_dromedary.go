package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steel Dromedary — Artifact Creature — Camel {3}, 2/2 (Modern Horizons 2,
// #5501):
//
//	"This creature enters tapped with two +1/+1 counters on it.
//	 This creature doesn't untap during your untap step if it has a +1/+1
//	 counter on it.
//	 At the beginning of combat on your turn, you may move a +1/+1 counter
//	 from this creature onto target creature."
//
// The entry is Baloth Prime's tapped-with-counters replacement. The untap
// restriction reads the counters as the untap step runs. The combat trigger
// targets as it goes on the stack and the "may" is asked as it resolves
// (CR 603.5); the counter moves (it is removed here and put on the target)
// only if the Dromedary still has one and the target is still legal, and a
// Dromedary that has left does nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40fa5656-354f-47f9-ba8d-40291ee5b40e",
		Name:         "Steel Dromedary",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedWithCounters(game.CounterPlusOne, 2)},
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label: "This creature doesn't untap during your untap step if it has a +1/+1 counter on it.",
			Restricts: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.Counters[game.CounterPlusOne] > 0
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.AtBeginningOfYourCombat("Steel Dromedary — move a +1/+1 counter from this creature onto target creature",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.SourceIsStillThisPermanent(g, item) {
							return nil
						}
						ctx := effects.NewContext(g, item)
						target := effects.FirstLegalBattlefieldTarget(ctx)
						src, ok := g.LookupCardForEffect(item.SourceCardID)
						if target == uuid.Nil || !ok || src.Counters[game.CounterPlusOne] < 1 {
							return nil
						}
						if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: -1}).Apply(ctx); err != nil {
							return err
						}
						return effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
				"Steel Dromedary — move a +1/+1 counter onto the target creature?"),
		},
	})
}
