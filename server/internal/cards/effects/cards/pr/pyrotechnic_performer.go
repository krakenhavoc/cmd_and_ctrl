package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrotechnic Performer — Creature — Lizard Assassin {1}{R}, 3/2
// (Murders at Karlov Manor, #4469):
//
//	"Disguise {R} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 Whenever this creature or another creature you control is turned
//	 face up, that creature deals damage equal to its power to each
//	 opponent."
//
// The Performer has the ability once it is face up, so turning it face
// up triggers it too (CR 708.8). The creature that turned face up is
// the damage's source, and its power is read as the trigger resolves,
// or as it last existed if it has left the battlefield (CR 608.2h). One
// damage instance from it to every opponent, as Origin of Thor deals it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "78f9ec0f-6015-4024-aca6-9a14cba373fa",
		Name:             "Pyrotechnic Performer",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTurnedFaceUp, effects.ACreatureYouControlWasTurnedFaceUp,
				"Pyrotechnic Performer — that creature deals damage equal to its power to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					creature := item.Trigger.Event.CardID
					power := effects.PowerNowOrLastKnown(g, creature)
					if power <= 0 {
						return nil
					}
					return effects.OriginOfThorDealToEachOpponent(effects.NewContext(g, item), creature, power)
				}),
		},
	})
}
