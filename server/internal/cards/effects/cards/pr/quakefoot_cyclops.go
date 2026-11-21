package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cbb6139f-827f-490b-b9b7-e8c3ef714d44",
		Name:         "Quakefoot Cyclops",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Quakefoot Cyclops — up to two target creatures can't block this turn",
					effects.GapB157TargetsCantBlockThisTurn("Quakefoot Cyclops — can't block")),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2)),
			effects.Targeting(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
					"Quakefoot Cyclops — cycled: target creature can't block this turn",
					effects.GapB157TargetsCantBlockThisTurn("Quakefoot Cyclops — can't block"))),
				effects.TargetCreature("target creature")),
		},
	})
}
