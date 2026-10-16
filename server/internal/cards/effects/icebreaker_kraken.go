package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "01590ff3-34fa-47aa-b5da-f0b8c1b031c6",
		Name:         "Icebreaker Kraken",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for snow lands", And(Land(), Snow())),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Icebreaker Kraken — artifacts and creatures that player controls don't untap during their next untap step",
				func(g *game.Game, item *game.StackItem) error {
					return targetPlayersPermanentsDontUntap(NewContext(g, item),
						"Icebreaker Kraken — artifacts and creatures don't untap during that player's next untap step",
						"artifact", "creature")
				}), TargetPlayer("target opponent", Opponent())),
		},
		Activated: []ActivatedAbility{{
			Label:   "Return three snow lands you control to their owner's hand: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ReturnNToHand(3, "three snow lands you control", Land(), Snow()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
