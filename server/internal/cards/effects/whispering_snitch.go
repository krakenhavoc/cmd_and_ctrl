package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b6836023-4475-415e-9671-980d9eb6e762",
		Name:         "Whispering Snitch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSurveil, cmdr2B27IsYourFirstSurveilThisTurn,
				"Whispering Snitch — 1 damage to each opponent, gain 1 life",
				Do(DamageEachOpponent{N: 1}, GainLife{Amount: 1})),
		},
	})
}
