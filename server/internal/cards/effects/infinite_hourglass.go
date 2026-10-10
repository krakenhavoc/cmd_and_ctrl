package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infinite Hourglass — Artifact {4}:
//
//	"At the beginning of your upkeep, put a time counter on this
//	 artifact.
//	 All creatures get +1/+0 for each time counter on this artifact.
//	 {3}: Remove a time counter from this artifact. Any player may
//	 activate this ability but only during any upkeep step."
//
// ADR 0106 PR 6 (#1793). Armageddon Clock's shape with a static in the
// middle:
//
//   - The upkeep trigger puts the counter on (AtYourUpkeep).
//   - "All creatures get +1/+0 for each time counter" is a layer-7c
//     modification (CR 613.4c) over every creature, every controller's,
//     read off the Hourglass's counters on every layer pass. A counter
//     placed or removed bumps the layer version (EventCounterPlaced),
//     so the bonus follows the count at once.
//   - The any-player row (CR 602.2, 602.1b): "only during any upkeep
//     step" is DuringStep(StepUpkeep), any player's upkeep. The {3} is
//     the activator's to pay (CR 602.1a). Removing the counter is the
//     effect, so with none left it does nothing (CR 609.3).
//
// A time counter on a permanent means nothing to suspend, which reads
// time counters only on a card in exile (CR 702.62b).
//
// No purpose for the bot: the effect changes every creature at the
// table and no Purpose field describes it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c59ad65-ad1f-4dc4-bcf2-ea6b2b48fad8",
		Name:         "Infinite Hourglass",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Infinite Hourglass — put a time counter", putACounterOnThis(game.CounterTime)),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.Power += source.Counters[game.CounterTime]
			},
		}},
		Activated: []ActivatedAbility{{
			Label:     "{3}: Remove a time counter from this artifact. Any player may activate this ability but only during any upkeep step.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      game.AbilityCost{Mana: "{3}"},
			AnyPlayer: true,
			Condition: DuringStep(game.StepUpkeep),
			Effect:    removeACounterFromThis(game.CounterTime),
		}},
	})
}
