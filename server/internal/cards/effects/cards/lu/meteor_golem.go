package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meteor Golem — Artifact Creature — Golem {7}, 3/3 (EDHREC rank 794):
//
//	"When this creature enters, destroy target nonland permanent an
//	 opponent controls."
//
// Colourless removal on a body — the reason it is in every deck that
// can't cast Beast Within, and a blink target. Acidic Slime's shape
// exactly: a targeted ETB trigger whose target the controller picks
// when it fires, removed if no opponent controls a nonland permanent
// (CR 603.3d). Mandatory, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9f11aa1-9219-42a8-85a9-a8f204160706",
		Name:         "Meteor Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets:   effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls()),
			Key:       "Meteor Golem — destroy target nonland permanent an opponent controls",
			Effect:    effects.DestroyChosenPermanent,
		}},
	})
}
