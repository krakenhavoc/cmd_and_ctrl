package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icebreaker Kraken — Snow Creature — Kraken {10}{U}{U}, 8/8:
//
//	"Affinity for snow lands (This spell costs {1} less to cast for each
//	 snow land you control.)
//	 When this creature enters, artifacts and creatures target opponent
//	 controls don't untap during that player's next untap step.
//	 Return three snow lands you control to their owner's hand: Return
//	 this creature to its owner's hand."
//
// Affinity is the CR 702.41a cost reduction over snow lands you control.
// The enters trigger is #3995's player-wide skip over two card types,
// read at that opponent's next untap step, so an artifact or creature
// they gain before then stays tapped too (CR 611.2c). The return of
// three snow lands is a cost (CR 602.2b), paid in full at activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01590ff3-34fa-47aa-b5da-f0b8c1b031c6",
		Name:         "Icebreaker Kraken",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for snow lands", effects.And(effects.Land(), effects.Snow())),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Icebreaker Kraken — artifacts and creatures that player controls don't untap during their next untap step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.TargetPlayersPermanentsDontUntap(effects.NewContext(g, item),
						"Icebreaker Kraken — artifacts and creatures don't untap during that player's next untap step",
						"artifact", "creature")
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Return three snow lands you control to their owner's hand: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ReturnNToHand(3, "three snow lands you control", effects.Land(), effects.Snow()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
