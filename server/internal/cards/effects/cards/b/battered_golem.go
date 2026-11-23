package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battered Golem — Artifact Creature — Golem {3}, 3/2 (Fifth Dawn,
// #5604):
//
//	"This creature doesn't untap during your untap step.
//	 Whenever an artifact enters, you may untap this creature."
//
// Brass Man's untap-step restriction and Grinding Station's optional
// untap, which watches every artifact entering under any player's
// control, the Golem's own entry included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "066dfefd-eef5-4c81-8e8a-310bac01674f",
		Name:                  "Battered Golem",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B20ArtifactEntered(ev, g)
			}, "Battered Golem — untap", effects.B31UntapSelf), "Battered Golem — untap it?"),
		},
	})
}
