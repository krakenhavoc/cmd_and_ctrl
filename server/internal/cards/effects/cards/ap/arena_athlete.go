package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arena Athlete — Creature — Human {1}{R}, 2/1 (Foundations Jumpstart,
// #5929):
//
//	"Heroic — Whenever you cast a spell that targets this creature, target
//	 creature an opponent controls can't block this turn."
//
// The shared heroic trigger (CR 207.2c) with a target of its own, chosen as
// the trigger goes on the stack; with no creature an opponent controls the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4dfea3bf-dfd1-4abd-a17e-f3a07488a069",
		Name:         "Arena Athlete",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Heroic("Arena Athlete — target creature an opponent controls can't block this turn (heroic)",
				effects.TargetCreatureCantBlockThisTurn("Arena Athlete — target creature can't block this turn")), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
