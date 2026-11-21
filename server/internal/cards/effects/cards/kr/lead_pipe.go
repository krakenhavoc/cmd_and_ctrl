package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lead Pipe — Artifact — Clue Equipment {B} (Murders at Karlov Manor,
// #4469):
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature dies, each opponent loses 1 life.
//	 {2}, Sacrifice this Equipment: Draw a card.
//	 Equip {2}"
//
// The dies trigger is Skullclamp's (equippedCreatureDied): it reads the
// attachment before the CR 704.5n unattach, and the Pipe's controller's
// opponents lose the life. The draw is Candlestick's, printed on the
// card, so sacrificing the Pipe is sacrificing a Clue for every card that
// watches that. Equip is CR 702.6a.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed2ef95e-8851-4226-9c63-4c19ea323b82",
		Name:         "Lead Pipe",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.EquippedCreatureDied(ev, source)
			}, "Lead Pipe — each opponent loses 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, Sacrifice this Equipment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			effects.EquipAbility("{2}"),
		},
	})
}
