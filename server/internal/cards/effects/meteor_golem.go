package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d9f11aa1-9219-42a8-85a9-a8f204160706",
		Name:         "Meteor Golem",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return destroyChosenTargetTrigger(source, "Meteor Golem — destroy target nonland permanent an opponent controls")
			},
		}},
	})
}
