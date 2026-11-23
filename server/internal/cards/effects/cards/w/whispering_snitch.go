package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whispering Snitch — Creature — Vampire Rogue {1}{B}, 1/3:
//
//	"Whenever you surveil for the first time each turn, this creature deals
//	 1 damage to each opponent and you gain 1 life."
//
// The surveil trigger gated on the controller's first surveil of the turn
// (cmdr2B27IsYourFirstSurveilThisTurn: no earlier surveil of theirs in the
// turn's log), then Vindictive Vampire's damage and gain. Damage, not life
// loss, so it can be prevented.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6836023-4475-415e-9671-980d9eb6e762",
		Name:         "Whispering Snitch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSurveil, effects.Cmdr2B27IsYourFirstSurveilThisTurn,
				"Whispering Snitch — 1 damage to each opponent, gain 1 life",
				effects.Do(effects.DamageEachOpponent{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
