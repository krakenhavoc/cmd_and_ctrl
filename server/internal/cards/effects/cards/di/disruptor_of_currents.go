package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disruptor of Currents — Creature — Merfolk Wizard {3}{U}{U}, 3/3
// (Lorwyn Eclipsed, #4458):
//
//	"Flash
//	 Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 When this creature enters, return up to one other target nonland
//	 permanent to its owner's hand."
//
// Deceit's {U}{U} trigger without the condition: "other" is object
// identity (Another), "up to one" lets it target nothing, and the
// target is checked again as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e76ae81f-729c-4def-ae83-6870a696cbe8",
		Name:            "Disruptor of Currents",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		TapCost:         effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Disruptor of Currents — return up to one other target nonland permanent to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.BounceTheTarget(item, effects.NewContext(g, item))
					}),
				effects.Another(effects.TargetPermanent("up to one other target nonland permanent", effects.Nonland()).WithCount(0, 1))),
		},
	})
}
