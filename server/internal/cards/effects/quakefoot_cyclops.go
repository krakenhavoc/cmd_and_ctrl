package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quakefoot Cyclops — Creature — Cyclops {4}{R}, 5/4 (Jumpstart 2022, #5948):
//
//	"When this creature enters, up to two target creatures can't block this
//	 turn.
//	 Cycling {1}{R}
//	 When you cycle this card, target creature can't block this turn."
//
// Markov Warlord's enters trigger plus Agonasaur Rex's cycled-from-graveyard
// trigger (CR 702.29c): the card is in the graveyard when it triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbb6139f-827f-490b-b9b7-e8c3ef714d44",
		Name:         "Quakefoot Cyclops",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Quakefoot Cyclops — up to two target creatures can't block this turn",
					gapB157TargetsCantBlockThisTurn("Quakefoot Cyclops — can't block")),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
			Targeting(
				InGraveyard(On(game.EventCycle, Self,
					"Quakefoot Cyclops — cycled: target creature can't block this turn",
					gapB157TargetsCantBlockThisTurn("Quakefoot Cyclops — can't block"))),
				TargetCreature("target creature")),
		},
	})
}
